package airtable

import (
	"context"
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the Airtable client.
func ExampleUsage() {
	// Initialize the client
	// Get your API key from: https://airtable.com/account
	// Get your base ID from: https://airtable.com/api (select your base)
	apiKey := "your_api_key_here"
	baseID := "your_base_id_here"

	client, err := NewClient(apiKey, baseID)
	if err != nil {
		log.Fatalf("Failed to create Airtable client: %v", err)
	}

	ctx := context.Background()
	tableName := "Books" // Replace with your table name

	// Example 1: List all records
	records, err := client.ListRecords(ctx, tableName, nil)
	if err != nil {
		log.Printf("Failed to list records: %v", err)
	} else {
		fmt.Printf("Found %d records\n", len(records))
		for _, record := range records {
			fmt.Printf("Record ID: %s, Fields: %+v\n", record.ID, record.Fields)
		}
	}

	// Example 2: List records with filters and sorting
	params := &ListParams{
		View:            "Grid view",
		FilterByFormula: `AND({Status}='Active', {Price}>100)`,
		Sort: []SortParam{
			{Field: "Created", Direction: "desc"},
		},
	}
	filteredRecords, err := client.ListRecords(ctx, tableName, params)
	if err != nil {
		log.Printf("Failed to list filtered records: %v", err)
	} else {
		fmt.Printf("Found %d filtered records\n", len(filteredRecords))
	}

	// Example 3: Get a single record by ID
	recordID := "recXXXXXXXXXXXXXX" // Replace with actual record ID
	record, err := client.GetRecord(ctx, tableName, recordID)
	if err != nil {
		log.Printf("Failed to get record: %v", err)
	} else {
		fmt.Printf("Record: %+v\n", record)
	}

	// Example 4: Create a new record
	newFields := map[string]interface{}{
		"Title":  "The Go Programming Language",
		"Author": "Alan A. A. Donovan",
		"Price":  45.99,
	}
	createdRecord, err := client.CreateRecord(ctx, tableName, newFields)
	if err != nil {
		log.Printf("Failed to create record: %v", err)
	} else {
		fmt.Printf("Created record with ID: %s\n", createdRecord.ID)
	}

	// Example 5: Update a record (full update)
	updateFields := map[string]interface{}{
		"Title":  "The Go Programming Language (Updated)",
		"Author": "Alan A. A. Donovan & Brian W. Kernighan",
		"Price":  49.99,
	}
	updatedRecord, err := client.UpdateRecord(ctx, tableName, recordID, updateFields)
	if err != nil {
		log.Printf("Failed to update record: %v", err)
	} else {
		fmt.Printf("Updated record: %+v\n", updatedRecord)
	}

	// Example 6: Partial update (only update specific fields)
	partialFields := map[string]interface{}{
		"Price": 39.99,
	}
	partiallyUpdatedRecord, err := client.UpdateRecordPartial(ctx, tableName, recordID, partialFields)
	if err != nil {
		log.Printf("Failed to partially update record: %v", err)
	} else {
		fmt.Printf("Partially updated record: %+v\n", partiallyUpdatedRecord)
	}

	// Example 7: Delete a record
	err = client.DeleteRecord(ctx, tableName, recordID)
	if err != nil {
		log.Printf("Failed to delete record: %v", err)
	} else {
		fmt.Println("Record deleted successfully")
	}

	// Example 8: Bulk delete records (up to 10 at a time)
	idsToDelete := []string{"rec1", "rec2", "rec3"}
	err = client.BulkDeleteRecords(ctx, tableName, idsToDelete)
	if err != nil {
		log.Printf("Failed to bulk delete records: %v", err)
	} else {
		fmt.Println("Records deleted successfully")
	}
}

