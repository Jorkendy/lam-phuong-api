package jobtype

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

// JobType represents a job type.
type JobType struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

// JobTypeResponseWrapper wraps JobType in the standard API response format for Swagger
// @Description Response containing a single job type
type JobTypeResponseWrapper struct {
	Success bool    `json:"success" example:"true"`
	Data    JobType `json:"data"`
	Message string  `json:"message" example:"Job type retrieved successfully"`
}

// JobTypesResponseWrapper wraps array of JobTypes in the standard API response format for Swagger
// @Description Response containing a list of job types
type JobTypesResponseWrapper struct {
	Success bool      `json:"success" example:"true"`
	Data    []JobType `json:"data"`
	Message string    `json:"message" example:"Job types retrieved successfully"`
}

// ToAirtableFields converts a JobType to Airtable fields format (for creation)
// Deprecated: Use ToAirtableFieldsForCreate() instead
func (jt *JobType) ToAirtableFields() map[string]interface{} {
	return jt.ToAirtableFieldsForCreate()
}

// FromAirtable maps an Airtable record to a JobType.
// The record should have an "id" field and a "fields" map.
func FromAirtable(record map[string]interface{}) (*JobType, error) {
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

	return &JobType{
		ID:     id,
		Name:   getStringField(fields, FieldName),
		Slug:   getStringField(fields, FieldSlug),
		Status: status,
	}, nil
}

