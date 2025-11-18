package jobcategory

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

// JobCategory represents a job category.
type JobCategory struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

// JobCategoryResponseWrapper wraps JobCategory in the standard API response format for Swagger
// @Description Response containing a single job category
type JobCategoryResponseWrapper struct {
	Success bool        `json:"success" example:"true"`
	Data    JobCategory `json:"data"`
	Message string      `json:"message" example:"Job category retrieved successfully"`
}

// JobCategoriesResponseWrapper wraps array of JobCategories in the standard API response format for Swagger
// @Description Response containing a list of job categories
type JobCategoriesResponseWrapper struct {
	Success bool          `json:"success" example:"true"`
	Data    []JobCategory `json:"data"`
	Message string        `json:"message" example:"Job categories retrieved successfully"`
}

// ToAirtableFields converts a JobCategory to Airtable fields format (for creation)
// Deprecated: Use ToAirtableFieldsForCreate() instead
func (jc *JobCategory) ToAirtableFields() map[string]interface{} {
	return jc.ToAirtableFieldsForCreate()
}

// FromAirtable maps an Airtable record to a JobCategory.
// The record should have an "id" field and a "fields" map.
func FromAirtable(record map[string]interface{}) (*JobCategory, error) {
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

	return &JobCategory{
		ID:     id,
		Name:   getStringField(fields, FieldName),
		Slug:   getStringField(fields, FieldSlug),
		Status: status,
	}, nil
}

