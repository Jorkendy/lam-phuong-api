package location

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

// Location represents a physical place served by the API.
type Location struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
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

// FromAirtable maps an Airtable record to a Location.
// The record should have an "id" field and a "fields" map.
func FromAirtable(record map[string]interface{}) (*Location, error) {
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

	return &Location{
		ID:     id,
		Name:   getStringField(fields, FieldName),
		Slug:   getStringField(fields, FieldSlug),
		Status: status,
	}, nil
}
