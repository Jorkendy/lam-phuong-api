package location

import (
	"context"
	"fmt"
	"log"
	"strings"

	"lam-phuong-api/internal/airtable"
)

// Repository defines behavior for storing and retrieving locations.
type Repository interface {
	List() []Location
	Create(ctx context.Context, location Location) (Location, error)
	Get(id string) (Location, bool)
	GetBySlug(slug string) (Location, bool)
	Update(ctx context.Context, id string, location Location) (Location, error)
	DeleteBySlug(slug string) bool
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

// List returns all locations from Airtable
func (r *AirtableRepository) List() []Location {
	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, nil)
	if err != nil {
		log.Printf("Failed to list locations from Airtable: %v", err)
		return []Location{} // Return empty slice on error
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

	return locations
}

// Create adds a new location to Airtable
func (r *AirtableRepository) Create(ctx context.Context, location Location) (Location, error) {
	// Save to Airtable
	airtableFields := location.ToAirtableFieldsForCreate()
	log.Printf("Attempting to save location to Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.CreateRecord(ctx, r.airtableTable, airtableFields)
	if err != nil {
		log.Printf("Failed to save location to Airtable: %v", err)
		log.Printf("Error details - Table: %s, Fields: %+v", r.airtableTable, airtableFields)
		return Location{}, fmt.Errorf("failed to create location in Airtable: %w", err)
	}

	// Update the created location with Airtable ID
	location.ID = airtableRecord.ID
	log.Printf("Location saved to Airtable successfully with ID: %s", airtableRecord.ID)
	return location, nil
}

// DeleteBySlug removes a location by its slug from Airtable
func (r *AirtableRepository) DeleteBySlug(slug string) bool {
	filterValue := escapeAirtableFormulaValue(slug)
	params := &airtable.ListParams{
		FilterByFormula: fmt.Sprintf("{%s} = '%s'", FieldSlug, filterValue),
	}

	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, params)
	if err != nil {
		log.Printf("Failed to query Airtable for slug %s: %v", slug, err)
		return false
	}

	if len(records) == 0 {
		return false
	}

	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}

	if err := r.airtableClient.BulkDeleteRecords(context.Background(), r.airtableTable, ids); err != nil {
		log.Printf("Failed to delete Airtable records for slug %s: %v", slug, err)
		return false
	}

	return true
}

// Get retrieves a location by ID from Airtable
func (r *AirtableRepository) Get(id string) (Location, bool) {
	record, err := r.airtableClient.GetRecord(context.Background(), r.airtableTable, id)
	if err != nil {
		log.Printf("Failed to get location from Airtable: %v", err)
		return Location{}, false
	}

	loc, err := mapAirtableRecord(record)
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return Location{}, false
	}

	return loc, true
}

// GetBySlug retrieves a location by slug from Airtable
func (r *AirtableRepository) GetBySlug(slug string) (Location, bool) {
	filterValue := escapeAirtableFormulaValue(slug)
	params := &airtable.ListParams{
		FilterByFormula: fmt.Sprintf("{%s} = '%s'", FieldSlug, filterValue),
	}

	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, params)
	if err != nil {
		log.Printf("Failed to query Airtable for slug %s: %v", slug, err)
		return Location{}, false
	}

	if len(records) == 0 {
		return Location{}, false
	}

	loc, err := mapAirtableRecord(records[0])
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return Location{}, false
	}

	return loc, true
}

// Update updates a location in Airtable
func (r *AirtableRepository) Update(ctx context.Context, id string, location Location) (Location, error) {
	airtableFields := location.ToAirtableFieldsForUpdate()
	log.Printf("Attempting to update location in Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.UpdateRecordPartial(ctx, r.airtableTable, id, airtableFields)
	if err != nil {
		log.Printf("Failed to update location in Airtable: %v", err)
		return Location{}, fmt.Errorf("failed to update location in Airtable: %w", err)
	}

	updated, err := mapAirtableRecord(airtableRecord)
	if err != nil {
		return Location{}, fmt.Errorf("failed to map updated location: %w", err)
	}

	log.Printf("Location updated in Airtable successfully with ID: %s", id)
	return updated, nil
}

func mapAirtableRecord(record airtable.Record) (Location, error) {
	status := getStringField(record.Fields, FieldStatus)
	if status == "" {
		status = StatusActive // Default to Active if not set
	}
	return Location{
		ID:     record.ID,
		Name:   getStringField(record.Fields, FieldName),
		Slug:   getStringField(record.Fields, FieldSlug),
		Status: status,
	}, nil
}

func escapeAirtableFormulaValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
