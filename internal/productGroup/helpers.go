package productgroup

import "time"

// ToAirtableFieldsForCreate converts a ProductGroup to Airtable fields format for creation
func (pg *ProductGroup) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:      pg.Name,
		FieldSlug:      pg.Slug,
		FieldCreatedAt: now,
		FieldUpdatedAt: now,
	}
}

// ToAirtableFieldsForUpdate converts a ProductGroup to Airtable fields format for update
func (pg *ProductGroup) ToAirtableFieldsForUpdate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:      pg.Name,
		FieldSlug:      pg.Slug,
		FieldUpdatedAt: now,
	}
}
