package jobcategory

import "time"

// ToAirtableFieldsForCreate converts a JobCategory to Airtable fields format for creation
func (jc *JobCategory) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:      jc.Name,
		FieldSlug:      jc.Slug,
		FieldCreatedAt: now,
		FieldUpdatedAt: now,
	}
}

// ToAirtableFieldsForUpdate converts a JobCategory to Airtable fields format for update
func (jc *JobCategory) ToAirtableFieldsForUpdate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:      jc.Name,
		FieldSlug:      jc.Slug,
		FieldUpdatedAt: now,
	}
}

