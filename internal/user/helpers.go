package user

import "time"

// ToAirtableFieldsForCreate converts a User to Airtable fields format for creation
func (u *User) ToAirtableFieldsForCreate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	fields := map[string]interface{}{
		FieldEmail:     u.Email,
		FieldPassword:  u.Password, // Already hashed
		FieldCreatedAt: now,
		FieldUpdatedAt: now,
	}
	if u.Role != "" {
		fields[FieldRole] = u.Role
	}
	return fields
}

// ToAirtableFieldsForUpdate converts a User to Airtable fields format for update
func (u *User) ToAirtableFieldsForUpdate() map[string]interface{} {
	now := time.Now().Format(time.RFC3339)
	fields := map[string]interface{}{
		FieldEmail:     u.Email,
		FieldUpdatedAt: now,
	}
	if u.Password != "" {
		fields[FieldPassword] = u.Password // Already hashed
	}
	if u.Role != "" {
		fields[FieldRole] = u.Role
	}
	return fields
}

