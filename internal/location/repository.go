package location

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"

	"lam-phuong-api/internal/airtable"
)

// Repository defines behavior for storing and retrieving locations.
type Repository interface {
	List() []Location
	Get(id string) (Location, bool)
	Create(ctx context.Context, location Location) (Location, error)
	Update(id string, location Location) (Location, bool)
	Delete(id string) bool
}

// InMemoryRepository stores locations in memory and is safe for concurrent access.
type InMemoryRepository struct {
	mu     sync.RWMutex
	data   map[string]Location
	nextID int
}

// NewInMemoryRepository creates an in-memory repository seeded with optional data.
func NewInMemoryRepository(seed []Location) *InMemoryRepository {
	repo := &InMemoryRepository{
		data:   make(map[string]Location),
		nextID: 1,
	}

	maxID := 0
	for _, l := range seed {
		repo.data[l.ID] = l
		if id, err := strconv.Atoi(l.ID); err == nil && id > maxID {
			maxID = id
		}
	}
	repo.nextID = maxID + 1

	return repo
}

// List returns all locations sorted by ID.
func (r *InMemoryRepository) List() []Location {
	r.mu.RLock()
	defer r.mu.RUnlock()

	locations := make([]Location, 0, len(r.data))
	for _, location := range r.data {
		locations = append(locations, location)
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].ID < locations[j].ID
	})

	return locations
}

// Get retrieves a location by ID.
func (r *InMemoryRepository) Get(id string) (Location, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	location, ok := r.data[id]
	return location, ok
}

// Create adds a new location and automatically assigns an ID.
// Note: ctx parameter is for interface compatibility but not used in in-memory implementation.
func (r *InMemoryRepository) Create(ctx context.Context, location Location) (Location, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	location.ID = strconv.Itoa(r.nextID)
	r.nextID++
	r.data[location.ID] = location

	return location, nil
}

// Update modifies an existing location record.
func (r *InMemoryRepository) Update(id string, location Location) (Location, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return Location{}, false
	}

	location.ID = id
	r.data[id] = location
	return location, true
}

// Delete removes a location by ID.
func (r *InMemoryRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return false
	}

	delete(r.data, id)
	return true
}

// AirtableRepository wraps a Repository and adds Airtable persistence.
type AirtableRepository struct {
	repo           Repository
	airtableClient *airtable.Client
	airtableTable  string
}

// NewAirtableRepository creates a repository that syncs to Airtable.
func NewAirtableRepository(repo Repository, airtableClient *airtable.Client, airtableTable string) *AirtableRepository {
	return &AirtableRepository{
		repo:           repo,
		airtableClient: airtableClient,
		airtableTable:  airtableTable,
	}
}

// List returns all locations from the underlying repository.
func (r *AirtableRepository) List() []Location {
	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, nil)
	if err != nil {
		log.Printf("Failed to list locations from Airtable: %v", err)
		return r.repo.List()
	}

	locations := make([]Location, 0, len(records))
	for _, record := range records {
		loc, err := mapAirtableRecord(record)
		if err != nil {
			log.Printf("Skipping Airtable record due to mapping error: %v", err)
			continue
		}
		locations = append(locations, loc)
	}

	// If Airtable returns no records, fall back to underlying repository
	if len(locations) == 0 {
		return r.repo.List()
	}

	return locations
}

// Get retrieves a location by ID from the underlying repository.
func (r *AirtableRepository) Get(id string) (Location, bool) {
	return r.repo.Get(id)
}

// Create adds a new location to the repository and syncs it to Airtable.
func (r *AirtableRepository) Create(ctx context.Context, location Location) (Location, error) {
	// Create in the underlying repository first
	created, err := r.repo.Create(ctx, location)
	if err != nil {
		return Location{}, err
	}

	// Save to Airtable
	airtableFields := created.ToAirtableFields()
	log.Printf("Attempting to save location to Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.CreateRecord(ctx, r.airtableTable, airtableFields)
	if err != nil {
		// Log error but don't fail - location is already created in repo
		log.Printf("Failed to save location to Airtable: %v", err)
		log.Printf("Error details - Table: %s, Fields: %+v", r.airtableTable, airtableFields)
		return created, nil // Return created location even if Airtable save failed
	}

	// Update the created location with Airtable ID
	created.ID = airtableRecord.ID
	log.Printf("Location saved to Airtable successfully with ID: %s", airtableRecord.ID)
	return created, nil
}

// Update modifies an existing location in the underlying repository.
func (r *AirtableRepository) Update(id string, location Location) (Location, bool) {
	return r.repo.Update(id, location)
}

// Delete removes a location from the underlying repository.
func (r *AirtableRepository) Delete(id string) bool {
	return r.repo.Delete(id)
}

func mapAirtableRecord(record airtable.Record) (Location, error) {
	status, err := getStatusFromFields(record.Fields)
	if err != nil {
		return Location{}, fmt.Errorf("invalid status for record %s: %w", record.ID, err)
	}

	return Location{
		ID:     record.ID,
		Name:   getStringField(record.Fields, FieldName),
		Slug:   getStringField(record.Fields, FieldSlug),
		Status: status,
	}, nil
}
