package jobtype

import (
	"context"
	"fmt"
	"log"
	"strings"

	"lam-phuong-api/internal/airtable"
)

// Repository defines behavior for storing and retrieving job types.
type Repository interface {
	List() []JobType
	Create(ctx context.Context, jobType JobType) (JobType, error)
	Get(id string) (JobType, bool)
	GetBySlug(slug string) (JobType, bool)
	Update(ctx context.Context, id string, jobType JobType) (JobType, error)
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

// List returns all job types from Airtable
func (r *AirtableRepository) List() []JobType {
	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, nil)
	if err != nil {
		log.Printf("Failed to list job types from Airtable: %v", err)
		return []JobType{} // Return empty slice on error
	}

	jobTypes := make([]JobType, 0, len(records))
	for _, record := range records {
		jt, err := mapAirtableRecord(record)
		if err != nil {
			log.Printf("Skipping Airtable record due to mapping error: %v", err)
			continue
		}
		jobTypes = append(jobTypes, jt)
	}

	return jobTypes
}

// Create adds a new job type to Airtable
func (r *AirtableRepository) Create(ctx context.Context, jobType JobType) (JobType, error) {
	// Save to Airtable
	airtableFields := jobType.ToAirtableFieldsForCreate()
	log.Printf("Attempting to save job type to Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.CreateRecord(ctx, r.airtableTable, airtableFields)
	if err != nil {
		log.Printf("Failed to save job type to Airtable: %v", err)
		log.Printf("Error details - Table: %s, Fields: %+v", r.airtableTable, airtableFields)
		return JobType{}, fmt.Errorf("failed to create job type in Airtable: %w", err)
	}

	// Update the created job type with Airtable ID
	jobType.ID = airtableRecord.ID
	log.Printf("Job type saved to Airtable successfully with ID: %s", airtableRecord.ID)
	return jobType, nil
}

// DeleteBySlug removes a job type by its slug from Airtable
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

// Get retrieves a job type by ID from Airtable
func (r *AirtableRepository) Get(id string) (JobType, bool) {
	record, err := r.airtableClient.GetRecord(context.Background(), r.airtableTable, id)
	if err != nil {
		log.Printf("Failed to get job type from Airtable: %v", err)
		return JobType{}, false
	}

	jt, err := mapAirtableRecord(record)
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return JobType{}, false
	}

	return jt, true
}

// GetBySlug retrieves a job type by slug from Airtable
func (r *AirtableRepository) GetBySlug(slug string) (JobType, bool) {
	filterValue := escapeAirtableFormulaValue(slug)
	params := &airtable.ListParams{
		FilterByFormula: fmt.Sprintf("{%s} = '%s'", FieldSlug, filterValue),
	}

	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, params)
	if err != nil {
		log.Printf("Failed to query Airtable for slug %s: %v", slug, err)
		return JobType{}, false
	}

	if len(records) == 0 {
		return JobType{}, false
	}

	jt, err := mapAirtableRecord(records[0])
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return JobType{}, false
	}

	return jt, true
}

// Update updates a job type in Airtable
func (r *AirtableRepository) Update(ctx context.Context, id string, jobType JobType) (JobType, error) {
	airtableFields := jobType.ToAirtableFieldsForUpdate()
	log.Printf("Attempting to update job type in Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.UpdateRecordPartial(ctx, r.airtableTable, id, airtableFields)
	if err != nil {
		log.Printf("Failed to update job type in Airtable: %v", err)
		return JobType{}, fmt.Errorf("failed to update job type in Airtable: %w", err)
	}

	updated, err := mapAirtableRecord(airtableRecord)
	if err != nil {
		return JobType{}, fmt.Errorf("failed to map updated job type: %w", err)
	}

	log.Printf("Job type updated in Airtable successfully with ID: %s", id)
	return updated, nil
}

func mapAirtableRecord(record airtable.Record) (JobType, error) {
	status := getStringField(record.Fields, FieldStatus)
	if status == "" {
		status = StatusActive // Default to Active if not set
	}
	return JobType{
		ID:     record.ID,
		Name:   getStringField(record.Fields, FieldName),
		Slug:   getStringField(record.Fields, FieldSlug),
		Status: status,
	}, nil
}

func escapeAirtableFormulaValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

