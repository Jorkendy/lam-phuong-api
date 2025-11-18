package location

import "time"

// ToAirtableFieldsForCreate converts a Location to Airtable fields format for creation
func (l *Location) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	status := l.Status
	if status == "" {
		status = StatusActive // Default to Active if not set
	}
	return map[string]interface{}{
		FieldName:      l.Name,
		FieldSlug:      l.Slug,
		FieldStatus:    status,
		FieldCreatedAt: now,
		FieldUpdatedAt: now,
	}
}

// ToAirtableFieldsForUpdate converts a Location to Airtable fields format for update
func (l *Location) ToAirtableFieldsForUpdate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:      l.Name,
		FieldSlug:      l.Slug,
		FieldStatus:    l.Status,
		FieldUpdatedAt: now,
	}
}

