package productgroup

import (
	"context"
	"fmt"
	"log"
	"strings"

	"lam-phuong-api/internal/airtable"
)

// Repository defines behavior for storing and retrieving product groups.
type Repository interface {
	List() []ProductGroup
	Create(ctx context.Context, productGroup ProductGroup) (ProductGroup, error)
	Get(id string) (ProductGroup, bool)
	GetBySlug(slug string) (ProductGroup, bool)
	Update(ctx context.Context, id string, productGroup ProductGroup) (ProductGroup, error)
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

// List returns all product groups from Airtable
func (r *AirtableRepository) List() []ProductGroup {
	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, nil)
	if err != nil {
		log.Printf("Failed to list product groups from Airtable: %v", err)
		return []ProductGroup{} // Return empty slice on error
	}

	productGroups := make([]ProductGroup, 0, len(records))
	for _, record := range records {
		pg, err := mapAirtableRecord(record)
		if err != nil {
			log.Printf("Skipping Airtable record due to mapping error: %v", err)
			continue
		}
		productGroups = append(productGroups, pg)
	}

	return productGroups
}

// Create adds a new product group to Airtable
func (r *AirtableRepository) Create(ctx context.Context, productGroup ProductGroup) (ProductGroup, error) {
	// Save to Airtable
	airtableFields := productGroup.ToAirtableFieldsForCreate()
	log.Printf("Attempting to save product group to Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.CreateRecord(ctx, r.airtableTable, airtableFields)
	if err != nil {
		log.Printf("Failed to save product group to Airtable: %v", err)
		log.Printf("Error details - Table: %s, Fields: %+v", r.airtableTable, airtableFields)
		return ProductGroup{}, fmt.Errorf("failed to create product group in Airtable: %w", err)
	}

	// Update the created product group with Airtable ID
	productGroup.ID = airtableRecord.ID
	log.Printf("Product group saved to Airtable successfully with ID: %s", airtableRecord.ID)
	return productGroup, nil
}

// DeleteBySlug removes a product group by its slug from Airtable
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

// Get retrieves a product group by ID from Airtable
func (r *AirtableRepository) Get(id string) (ProductGroup, bool) {
	record, err := r.airtableClient.GetRecord(context.Background(), r.airtableTable, id)
	if err != nil {
		log.Printf("Failed to get product group from Airtable: %v", err)
		return ProductGroup{}, false
	}

	pg, err := mapAirtableRecord(record)
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return ProductGroup{}, false
	}

	return pg, true
}

// GetBySlug retrieves a product group by slug from Airtable
func (r *AirtableRepository) GetBySlug(slug string) (ProductGroup, bool) {
	filterValue := escapeAirtableFormulaValue(slug)
	params := &airtable.ListParams{
		FilterByFormula: fmt.Sprintf("{%s} = '%s'", FieldSlug, filterValue),
	}

	records, err := r.airtableClient.ListRecords(context.Background(), r.airtableTable, params)
	if err != nil {
		log.Printf("Failed to query Airtable for slug %s: %v", slug, err)
		return ProductGroup{}, false
	}

	if len(records) == 0 {
		return ProductGroup{}, false
	}

	pg, err := mapAirtableRecord(records[0])
	if err != nil {
		log.Printf("Failed to map Airtable record: %v", err)
		return ProductGroup{}, false
	}

	return pg, true
}

// Update updates a product group in Airtable
func (r *AirtableRepository) Update(ctx context.Context, id string, productGroup ProductGroup) (ProductGroup, error) {
	airtableFields := productGroup.ToAirtableFieldsForUpdate()
	log.Printf("Attempting to update product group in Airtable table: %s", r.airtableTable)
	airtableRecord, err := r.airtableClient.UpdateRecordPartial(ctx, r.airtableTable, id, airtableFields)
	if err != nil {
		log.Printf("Failed to update product group in Airtable: %v", err)
		return ProductGroup{}, fmt.Errorf("failed to update product group in Airtable: %w", err)
	}

	updated, err := mapAirtableRecord(airtableRecord)
	if err != nil {
		return ProductGroup{}, fmt.Errorf("failed to map updated product group: %w", err)
	}

	log.Printf("Product group updated in Airtable successfully with ID: %s", id)
	return updated, nil
}

func mapAirtableRecord(record airtable.Record) (ProductGroup, error) {
	status := getStringField(record.Fields, FieldStatus)
	if status == "" {
		status = StatusActive // Default to Active if not set
	}
	return ProductGroup{
		ID:     record.ID,
		Name:   getStringField(record.Fields, FieldName),
		Slug:   getStringField(record.Fields, FieldSlug),
		Status: status,
	}, nil
}

func escapeAirtableFormulaValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

