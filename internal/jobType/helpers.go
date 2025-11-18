package jobtype

import "time"

// ToAirtableFieldsForCreate converts a JobType to Airtable fields format for creation
func (jt *JobType) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	status := jt.Status
	if status == "" {
		status = StatusActive // Default to Active if not set
	}
	return map[string]interface{}{
		FieldName:      jt.Name,
		FieldSlug:      jt.Slug,
		FieldStatus:    status,
		FieldCreatedAt: now,
		FieldUpdatedAt: now,
	}
}

// ToAirtableFieldsForUpdate converts a JobType to Airtable fields format for update
func (jt *JobType) ToAirtableFieldsForUpdate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{
		FieldName:      jt.Name,
		FieldSlug:      jt.Slug,
		FieldStatus:    jt.Status,
		FieldUpdatedAt: now,
	}
}

