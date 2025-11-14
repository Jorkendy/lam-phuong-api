package location

import "time"

// ToAirtableFieldsForCreate converts a Location to Airtable fields format for creation
func (l *Location) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:     l.Name,
		FieldSlug:     l.Slug,
		FieldCreatedAt: now,
		FieldUpdatedAt: now,
	}
}

// ToAirtableFieldsForUpdate converts a Location to Airtable fields format for update
func (l *Location) ToAirtableFieldsForUpdate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:     l.Name,
		FieldSlug:     l.Slug,
		FieldUpdatedAt: now,
	}
}

