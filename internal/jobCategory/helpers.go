package jobcategory

import "time"

// ToAirtableFieldsForCreate converts a JobCategory to Airtable fields format for creation
func (jc *JobCategory) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	status := jc.Status
	if status == "" {
		status = StatusActive // Default to Active if not set
	}
	return map[string]interface{}{
		FieldName:      jc.Name,
		FieldSlug:      jc.Slug,
		FieldStatus:    status,
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
		FieldStatus:    jc.Status,
		FieldUpdatedAt: now,
	}
}

