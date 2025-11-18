package location

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

// Location represents a physical place served by the API.
type Location struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// LocationResponseWrapper wraps Location in the standard API response format for Swagger
// @Description Response containing a single location
type LocationResponseWrapper struct {
	Success bool     `json:"success" example:"true"`
	Data    Location `json:"data"`
	Message string   `json:"message" example:"Location retrieved successfully"`
}

// LocationsResponseWrapper wraps array of Locations in the standard API response format for Swagger
// @Description Response containing a list of locations
type LocationsResponseWrapper struct {
	Success bool       `json:"success" example:"true"`
	Data    []Location `json:"data"`
	Message string     `json:"message" example:"Locations retrieved successfully"`
}

// ToAirtableFields converts a Location to Airtable fields format (for creation)
// Deprecated: Use ToAirtableFieldsForCreate() instead
func (l *Location) ToAirtableFields() map[string]interface{} {
	return l.ToAirtableFieldsForCreate()
}
