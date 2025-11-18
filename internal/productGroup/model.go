package productgroup

import "fmt"

// Airtable field names
const (
	FieldName      = "Name"
	FieldSlug      = "Slug"
	FieldStatus    = "Status"
	FieldCreatedAt = "Created At"
	FieldUpdatedAt = "Updated At"
)

// Status constants
const (
	StatusActive   = "Active"
	StatusDisabled = "Disabled"
)

// Helper functions
func getStringField(fields map[string]interface{}, key string) string {
	if val, ok := fields[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ProductGroup represents a product group/category.
type ProductGroup struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

// ProductGroupResponseWrapper wraps ProductGroup in the standard API response format for Swagger
// @Description Response containing a single product group
type ProductGroupResponseWrapper struct {
	Success bool         `json:"success" example:"true"`
	Data    ProductGroup `json:"data"`
	Message string       `json:"message" example:"Product group retrieved successfully"`
}

// ProductGroupsResponseWrapper wraps array of ProductGroups in the standard API response format for Swagger
// @Description Response containing a list of product groups
type ProductGroupsResponseWrapper struct {
	Success bool           `json:"success" example:"true"`
	Data    []ProductGroup `json:"data"`
	Message string         `json:"message" example:"Product groups retrieved successfully"`
}

// ToAirtableFields converts a ProductGroup to Airtable fields format (for creation)
// Deprecated: Use ToAirtableFieldsForCreate() instead
func (pg *ProductGroup) ToAirtableFields() map[string]interface{} {
	return pg.ToAirtableFieldsForCreate()
}

// FromAirtable maps an Airtable record to a ProductGroup.
// The record should have an "id" field and a "fields" map.
func FromAirtable(record map[string]interface{}) (*ProductGroup, error) {
	// Safely extract ID
	id := ""
	if idVal, ok := record["id"]; ok {
		if idStr, ok := idVal.(string); ok {
			id = idStr
		}
	}

	// Safely extract fields
	fields, ok := record["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid record: missing or invalid 'fields'")
	}

	status := getStringField(fields, FieldStatus)
	if status == "" {
		status = StatusActive // Default to Active if not set
	}

	return &ProductGroup{
		ID:     id,
		Name:   getStringField(fields, FieldName),
		Slug:   getStringField(fields, FieldSlug),
		Status: status,
	}, nil
}
