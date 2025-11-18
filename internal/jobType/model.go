package jobtype

// Airtable field names
const (
	FieldName      = "Name"
	FieldSlug      = "Slug"
	FieldCreatedAt = "Created At"
	FieldUpdatedAt = "Updated At"
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
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
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

