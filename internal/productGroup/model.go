package productgroup

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

// ProductGroup represents a product group/category.
type ProductGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
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
