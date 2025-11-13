package airtable

import (
	"context"
	"fmt"

	"github.com/mehanizm/airtable"
)

// Client wraps the mehanizm/airtable client with a simplified interface.
type Client struct {
	client *airtable.Client
	baseID string
}

// NewClient creates a new Airtable client using the mehanizm/airtable library.
// apiKey: Your Airtable API token (get from https://airtable.com/account)
// baseID: Your Airtable base ID (found in the API documentation for your base)
func NewClient(apiKey, baseID string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("airtable: api key is required")
	}
	if baseID == "" {
		return nil, fmt.Errorf("airtable: base ID is required")
	}

	client := airtable.NewClient(apiKey)

	return &Client{
		client: client,
		baseID: baseID,
	}, nil
}

// Record represents an Airtable record with simplified structure.
type Record struct {
	ID          string
	Fields      map[string]interface{}
	CreatedTime string
}

// ListParams configures ListRecords queries.
type ListParams struct {
	View            string
	PageSize        int
	FilterByFormula string
	Sort            []SortParam
}

// SortParam configures sorting for list queries.
type SortParam struct {
	Field     string
	Direction string // "asc" or "desc"
}

// ListRecords retrieves records from the specified table.
func (c *Client) ListRecords(ctx context.Context, table string, params *ListParams) ([]Record, error) {
	airtableTable := c.client.GetTable(c.baseID, table)

	query := airtableTable.GetRecords()

	if params != nil {
		if params.View != "" {
			query = query.FromView(params.View)
		}
		if params.FilterByFormula != "" {
			query = query.WithFilterFormula(params.FilterByFormula)
		}
		if len(params.Sort) > 0 {
			sortQueries := make([]struct {
				FieldName string
				Direction string
			}, 0, len(params.Sort))
			for _, sort := range params.Sort {
				direction := "asc"
				if sort.Direction == "desc" {
					direction = "desc"
				}
				sortQueries = append(sortQueries, struct {
					FieldName string
					Direction string
				}{
					FieldName: sort.Field,
					Direction: direction,
				})
			}
			query = query.WithSort(sortQueries...)
		}
	}

	var records *airtable.Records
	var err error
	if ctx != nil && ctx != context.Background() {
		records, err = query.DoContext(ctx)
	} else {
		records, err = query.Do()
	}
	if err != nil {
		return nil, fmt.Errorf("airtable: list records failed: %w", err)
	}

	result := make([]Record, 0, len(records.Records))
	for _, r := range records.Records {
		result = append(result, Record{
			ID:          r.ID,
			Fields:      r.Fields,
			CreatedTime: r.CreatedTime,
		})
	}

	return result, nil
}

// GetRecord fetches a single record by ID.
func (c *Client) GetRecord(ctx context.Context, table, id string) (Record, error) {
	airtableTable := c.client.GetTable(c.baseID, table)

	var record *airtable.Record
	var err error
	if ctx != nil && ctx != context.Background() {
		record, err = airtableTable.GetRecordContext(ctx, id)
	} else {
		record, err = airtableTable.GetRecord(id)
	}
	if err != nil {
		return Record{}, fmt.Errorf("airtable: get record failed: %w", err)
	}

	return Record{
		ID:          record.ID,
		Fields:      record.Fields,
		CreatedTime: record.CreatedTime,
	}, nil
}

// CreateRecord inserts a new record into Airtable.
func (c *Client) CreateRecord(ctx context.Context, table string, fields map[string]interface{}) (Record, error) {
	airtableTable := c.client.GetTable(c.baseID, table)

	recordsToSend := &airtable.Records{
		Records: []*airtable.Record{
			{
				Fields: fields,
			},
		},
	}

	receivedRecords, err := airtableTable.AddRecords(recordsToSend)
	if err != nil {
		return Record{}, fmt.Errorf("airtable: create record failed: %w", err)
	}

	if len(receivedRecords.Records) == 0 {
		return Record{}, fmt.Errorf("airtable: no record returned")
	}

	r := receivedRecords.Records[0]
	return Record{
		ID:          r.ID,
		Fields:      r.Fields,
		CreatedTime: r.CreatedTime,
	}, nil
}

// UpdateRecord replaces a record in Airtable (full update).
func (c *Client) UpdateRecord(ctx context.Context, table, id string, fields map[string]interface{}) (Record, error) {
	airtableTable := c.client.GetTable(c.baseID, table)

	// First get the record to update
	record, err := airtableTable.GetRecord(id)
	if err != nil {
		return Record{}, fmt.Errorf("airtable: get record for update failed: %w", err)
	}

	// Update all fields
	record.Fields = fields

	recordsToUpdate := &airtable.Records{
		Records: []*airtable.Record{record},
	}

	updatedRecords, err := airtableTable.UpdateRecords(recordsToUpdate)
	if err != nil {
		return Record{}, fmt.Errorf("airtable: update record failed: %w", err)
	}

	if len(updatedRecords.Records) == 0 {
		return Record{}, fmt.Errorf("airtable: no record returned")
	}

	r := updatedRecords.Records[0]
	return Record{
		ID:          r.ID,
		Fields:      r.Fields,
		CreatedTime: r.CreatedTime,
	}, nil
}

// UpdateRecordPartial performs a partial update on a record (only specified fields).
func (c *Client) UpdateRecordPartial(ctx context.Context, table, id string, fields map[string]interface{}) (Record, error) {
	airtableTable := c.client.GetTable(c.baseID, table)

	// First get the record to update
	var record *airtable.Record
	var err error
	if ctx != nil && ctx != context.Background() {
		record, err = airtableTable.GetRecordContext(ctx, id)
	} else {
		record, err = airtableTable.GetRecord(id)
	}
	if err != nil {
		return Record{}, fmt.Errorf("airtable: get record for update failed: %w", err)
	}

	// Convert fields to map[string]any for the library
	fieldsAny := make(map[string]any, len(fields))
	for k, v := range fields {
		fieldsAny[k] = v
	}

	// Use the library's UpdateRecordPartial method
	var updatedRecord *airtable.Record
	if ctx != nil && ctx != context.Background() {
		updatedRecord, err = record.UpdateRecordPartialContext(ctx, fieldsAny)
	} else {
		updatedRecord, err = record.UpdateRecordPartial(fieldsAny)
	}
	if err != nil {
		return Record{}, fmt.Errorf("airtable: partial update record failed: %w", err)
	}

	return Record{
		ID:          updatedRecord.ID,
		Fields:      updatedRecord.Fields,
		CreatedTime: updatedRecord.CreatedTime,
	}, nil
}

// DeleteRecord removes a record from Airtable.
func (c *Client) DeleteRecord(ctx context.Context, table, id string) error {
	airtableTable := c.client.GetTable(c.baseID, table)

	// Get the record first
	var record *airtable.Record
	var err error
	if ctx != nil && ctx != context.Background() {
		record, err = airtableTable.GetRecordContext(ctx, id)
	} else {
		record, err = airtableTable.GetRecord(id)
	}
	if err != nil {
		return fmt.Errorf("airtable: get record for delete failed: %w", err)
	}

	// Delete using the record's method
	if ctx != nil && ctx != context.Background() {
		_, err = record.DeleteRecordContext(ctx)
	} else {
		_, err = record.DeleteRecord()
	}
	if err != nil {
		return fmt.Errorf("airtable: delete record failed: %w", err)
	}

	return nil
}

// BulkDeleteRecords deletes multiple records (up to 10 at a time per Airtable API limits).
func (c *Client) BulkDeleteRecords(ctx context.Context, table string, ids []string) error {
	airtableTable := c.client.GetTable(c.baseID, table)

	_, err := airtableTable.DeleteRecords(ids)
	if err != nil {
		return fmt.Errorf("airtable: bulk delete records failed: %w", err)
	}

	return nil
}
