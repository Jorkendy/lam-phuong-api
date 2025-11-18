package jobcategory

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

// JobCategory represents a job category.
type JobCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
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
