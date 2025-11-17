package user

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

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
	GetByVerificationToken(token string) (User, bool)
}

// InMemoryRepository stores users in memory and is safe for concurrent access
type InMemoryRepository struct {
	mu     sync.RWMutex
	data   map[string]User
	nextID int
}

// NewInMemoryRepository creates a new in-memory user repository
func NewInMemoryRepository(seed []User) *InMemoryRepository {
	repo := &InMemoryRepository{
		data:   make(map[string]User),
		nextID: 1,
	}

	for _, u := range seed {
		repo.data[u.ID] = u
		if id, err := strconv.Atoi(u.ID); err == nil && id >= repo.nextID {
			repo.nextID = id + 1
		}
	}

	return repo
}

// List returns all users sorted by ID
func (r *InMemoryRepository) List() []User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]User, 0, len(r.data))
	for _, user := range r.data {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return users
}

// Create adds a new user and automatically assigns an ID
func (r *InMemoryRepository) Create(ctx context.Context, user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	for _, u := range r.data {
		if u.Email == user.Email {
			return User{}, fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	user.ID = strconv.Itoa(r.nextID)
	r.nextID++
	r.data[user.ID] = user

	return user, nil
}

// Delete removes a user by ID
func (r *InMemoryRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return false
	}

	delete(r.data, id)
	return true
}

// Get retrieves a user by ID
func (r *InMemoryRepository) Get(id string) (User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.data[id]
	return user, exists
}

// GetByEmail retrieves a user by email
func (r *InMemoryRepository) GetByEmail(email string) (User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.data {
		if user.Email == email {
			return user, true
		}
	}

	return User{}, false
}

// GetByVerificationToken retrieves a user by email verification token
func (r *InMemoryRepository) GetByVerificationToken(token string) (User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.data {
		if user.EmailVerificationToken == token {
			return user, true
		}
	}

	return User{}, false
}

// Update updates an existing user
func (r *InMemoryRepository) Update(ctx context.Context, id string, updatedUser User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	existingUser, exists := r.data[id]
	if !exists {
		return User{}, fmt.Errorf("user with id %s not found", id)
	}

	// Preserve ID and email (email should not be changed via update)
	updatedUser.ID = id
	updatedUser.Email = existingUser.Email

	// If password is empty, keep the existing password
	if updatedUser.Password == "" {
		updatedUser.Password = existingUser.Password
	}

	// If role is empty, keep the existing role
	if updatedUser.Role == "" {
		updatedUser.Role = existingUser.Role
	}

	r.data[id] = updatedUser
	return updatedUser, nil
}

// AirtableRepository wraps a Repository and adds Airtable persistence
type AirtableRepository struct {
	repo           Repository
	airtableClient *airtable.Client
	airtableTable  string
}

// NewAirtableRepository creates a repository that syncs to Airtable
func NewAirtableRepository(repo Repository, airtableClient *airtable.Client, airtableTable string) *AirtableRepository {
	return &AirtableRepository{
		repo:           repo,
		airtableClient: airtableClient,
		airtableTable:  airtableTable,
	}
}

// List returns all users from Airtable, falling back to underlying repository
func (r *AirtableRepository) List() []User {
	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, nil)
	if err != nil {
		log.Printf("Failed to list users from Airtable: %v", err)
		return r.repo.List()
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

	// If Airtable returns no records, fall back to underlying repository
	if len(users) == 0 {
		return r.repo.List()
	}

	return users
}

// Create adds a new user to the repository and syncs it to Airtable
func (r *AirtableRepository) Create(ctx context.Context, user User) (User, error) {
	// Create in the underlying repository first
	created, err := r.repo.Create(ctx, user)
	if err != nil {
		return User{}, err
	}

	// Save to Airtable
	airtableFields := created.ToAirtableFieldsForCreate()
	log.Printf("Attempting to save user to Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.CreateRecord(ctx, r.airtableTable, airtableFields)
	if err != nil {
		// Log error but don't fail - user is already created in repo
		log.Printf("Failed to save user to Airtable: %v", err)
		log.Printf("Error details - Table: %s, Fields: %+v", r.airtableTable, airtableFields)
		return created, nil // Return created user even if Airtable save failed
	}

	// Update the created user with Airtable ID
	created.ID = airtableRecord.ID
	log.Printf("User saved to Airtable successfully with ID: %s", airtableRecord.ID)
	return created, nil
}

// Delete removes a user from the underlying repository and Airtable
func (r *AirtableRepository) Delete(id string) bool {
	// Delete from underlying repository
	deleted := r.repo.Delete(id)
	if !deleted {
		return false
	}

	// Attempt to delete from Airtable
	if err := r.airtableClient.DeleteRecord(context.Background(), r.airtableTable, id); err != nil {
		log.Printf("Failed to delete Airtable record for user %s: %v", id, err)
	}

	return true
}

// Get retrieves a user by ID from Airtable, falling back to underlying repository
func (r *AirtableRepository) Get(id string) (User, bool) {
	record, err := r.airtableClient.GetRecord(context.Background(), r.airtableTable, id)
	if err != nil {
		log.Printf("Failed to get user from Airtable: %v", err)
		return r.repo.Get(id)
	}

	user, err := mapAirtableRecord(record)
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return r.repo.Get(id)
	}

	return user, true
}

// GetByEmail retrieves a user by email, preferring Airtable and falling back to repo cache
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
		return r.repo.GetByEmail(email)
	}

	if len(records) > 0 {
		user, mapErr := mapAirtableRecord(records[0])
		if mapErr == nil {
			return user, true
		}
		log.Printf("Failed to map Airtable user for email %s: %v", email, mapErr)
	}

	return r.repo.GetByEmail(email)
}

// GetByVerificationToken retrieves a user by verification token, preferring Airtable and falling back to repo cache
func (r *AirtableRepository) GetByVerificationToken(token string) (User, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, false
	}

	filter := fmt.Sprintf(
		"{%s} = '%s'",
		FieldEmailVerificationToken,
		escapeAirtableFormulaValue(token),
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
		log.Printf("Failed to find user by verification token in Airtable: %v", err)
		return r.repo.GetByVerificationToken(token)
	}

	if len(records) > 0 {
		user, mapErr := mapAirtableRecord(records[0])
		if mapErr == nil {
			return user, true
		}
		log.Printf("Failed to map Airtable user for token: %v", mapErr)
	}

	return r.repo.GetByVerificationToken(token)
}

// Update updates an existing user in the repository and syncs it to Airtable
func (r *AirtableRepository) Update(ctx context.Context, id string, updatedUser User) (User, error) {
	// Get existing user to preserve email
	existingUser, exists := r.repo.Get(id)
	if !exists {
		// Try to get from Airtable
		existingUser, exists = r.Get(id)
		if !exists {
			return User{}, fmt.Errorf("user with id %s not found", id)
		}
	}

	// Preserve email (should not be changed via update)
	updatedUser.Email = existingUser.Email

	// Update in the underlying repository first
	updated, err := r.repo.Update(ctx, id, updatedUser)
	if err != nil {
		return User{}, err
	}

	// Update in Airtable (partial update - only changed fields)
	airtableFields := updated.ToAirtableFieldsForUpdate()
	log.Printf("Attempting to update user in Airtable table: %s", r.airtableTable)
	_, err = r.airtableClient.UpdateRecordPartial(ctx, r.airtableTable, id, airtableFields)
	if err != nil {
		// Log error but don't fail - user is already updated in repo
		log.Printf("Failed to update user in Airtable: %v", err)
		log.Printf("Error details - Table: %s, ID: %s, Fields: %+v", r.airtableTable, id, airtableFields)
		return updated, nil // Return updated user even if Airtable update failed
	}

	log.Printf("User updated in Airtable successfully with ID: %s", id)
	return updated, nil
}

func mapAirtableRecord(record airtable.Record) (User, error) {
	role := getStringField(record.Fields, FieldRole)
	if role == "" {
		role = RoleUser // Default role
	}
	status := getStringField(record.Fields, FieldStatus)
	if status == "" {
		status = StatusPending // Default to pending
	}
	return User{
		ID:                    record.ID,
		Email:                 getStringField(record.Fields, FieldEmail),
		Password:              getStringField(record.Fields, FieldPassword),
		Role:                  role,
		Status:                status,
		EmailVerificationToken: getStringField(record.Fields, FieldEmailVerificationToken),
	}, nil
}

func escapeAirtableFormulaValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
