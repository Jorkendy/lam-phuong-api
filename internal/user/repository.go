package user

import (
	"context"
	"fmt"
	"log"
	"strings"

	"lam-phuong-api/internal/airtable"
)

// Repository defines behavior for storing and retrieving users
type Repository interface {
	List() []User
	Get(id string) (User, bool)
	Create(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, id string, user User) (User, error)
	Delete(id string) bool
	GetByEmail(email string) (User, bool)
}

// AirtableRepository implements Repository interface using Airtable as the data store
type AirtableRepository struct {
	airtableClient *airtable.Client
	airtableTable  string
}

// NewAirtableRepository creates a repository that uses Airtable as the data store
func NewAirtableRepository(airtableClient *airtable.Client, airtableTable string) *AirtableRepository {
	return &AirtableRepository{
		airtableClient: airtableClient,
		airtableTable:  airtableTable,
	}
}

// List returns all users from Airtable
func (r *AirtableRepository) List() []User {
	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, nil)
	if err != nil {
		log.Printf("Failed to list users from Airtable: %v", err)
		return []User{} // Return empty slice on error
	}

	users := make([]User, 0, len(records))
	for _, record := range records {
		user, err := mapAirtableRecord(record)
		if err != nil {
			log.Printf("Failed to map Airtable record: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users
}

// Create adds a new user to Airtable
func (r *AirtableRepository) Create(ctx context.Context, user User) (User, error) {
	// Check if email already exists
	_, exists := r.GetByEmail(user.Email)
	if exists {
		return User{}, fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Save to Airtable
	airtableFields := user.ToAirtableFieldsForCreate()
	log.Printf("Attempting to save user to Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.CreateRecord(ctx, r.airtableTable, airtableFields)
	if err != nil {
		log.Printf("Failed to save user to Airtable: %v", err)
		log.Printf("Error details - Table: %s, Fields: %+v", r.airtableTable, airtableFields)
		return User{}, fmt.Errorf("failed to create user in Airtable: %w", err)
	}

	// Update the created user with Airtable ID
	user.ID = airtableRecord.ID
	log.Printf("User saved to Airtable successfully with ID: %s", airtableRecord.ID)
	return user, nil
}

// Delete removes a user from Airtable
func (r *AirtableRepository) Delete(id string) bool {
	if err := r.airtableClient.DeleteRecord(context.Background(), r.airtableTable, id); err != nil {
		log.Printf("Failed to delete Airtable record for user %s: %v", id, err)
		return false
	}
	return true
}

// Get retrieves a user by ID from Airtable
func (r *AirtableRepository) Get(id string) (User, bool) {
	record, err := r.airtableClient.GetRecord(context.Background(), r.airtableTable, id)
	if err != nil {
		log.Printf("Failed to get user from Airtable: %v", err)
		return User{}, false
	}

	user, err := mapAirtableRecord(record)
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return User{}, false
	}

	return user, true
}

// GetByEmail retrieves a user by email from Airtable
func (r *AirtableRepository) GetByEmail(email string) (User, bool) {
	email = strings.TrimSpace(email)
	if email == "" {
		return User{}, false
	}

	filter := fmt.Sprintf(
		"LOWER({%s}) = '%s'",
		FieldEmail,
		escapeAirtableFormulaValue(strings.ToLower(email)),
	)

	records, err := r.airtableClient.ListRecords(
		context.Background(),
		r.airtableTable,
		&airtable.ListParams{
			PageSize:        1,
			FilterByFormula: filter,
		},
	)
	if err != nil {
		log.Printf("Failed to find user by email in Airtable: %v", err)
		return User{}, false
	}

	if len(records) > 0 {
		user, mapErr := mapAirtableRecord(records[0])
		if mapErr == nil {
			return user, true
		}
		log.Printf("Failed to map Airtable user for email %s: %v", email, mapErr)
	}

	return User{}, false
}


// Update updates an existing user in Airtable
func (r *AirtableRepository) Update(ctx context.Context, id string, updatedUser User) (User, error) {
	// Get existing user to preserve email
	existingUser, exists := r.Get(id)
	if !exists {
		return User{}, fmt.Errorf("user with id %s not found", id)
	}

	// Preserve email (should not be changed via update)
	updatedUser.Email = existingUser.Email
	updatedUser.ID = id

	// Preserve password if not provided
	if updatedUser.Password == "" {
		updatedUser.Password = existingUser.Password
	}

	// Preserve role if not provided
	if updatedUser.Role == "" {
		updatedUser.Role = existingUser.Role
	}

	// Preserve status if not provided
	if updatedUser.Status == "" {
		updatedUser.Status = existingUser.Status
	}

	// Update in Airtable (partial update - only changed fields)
	airtableFields := updatedUser.ToAirtableFieldsForUpdate()
	log.Printf("Attempting to update user in Airtable table: %s", r.airtableTable)
	_, err := r.airtableClient.UpdateRecordPartial(ctx, r.airtableTable, id, airtableFields)
	if err != nil {
		log.Printf("Failed to update user in Airtable: %v", err)
		log.Printf("Error details - Table: %s, ID: %s, Fields: %+v", r.airtableTable, id, airtableFields)
		return User{}, fmt.Errorf("failed to update user in Airtable: %w", err)
	}

	log.Printf("User updated in Airtable successfully with ID: %s", id)
	return updatedUser, nil
}

func mapAirtableRecord(record airtable.Record) (User, error) {
	role := getStringField(record.Fields, FieldRole)
	if role == "" {
		role = RoleUser // Default role
	}
	status := getStringField(record.Fields, FieldStatus)
	if status == "" {
		status = StatusActive // Default to active
	}
	return User{
		ID:       record.ID,
		Email:    getStringField(record.Fields, FieldEmail),
		Password: getStringField(record.Fields, FieldPassword),
		Role:     role,
		Status:   status,
	}, nil
}

func escapeAirtableFormulaValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
