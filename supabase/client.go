package supabase

import (
	"errors"
	"fmt"
	"os"

	"github.com/nedpals/supabase-go"
)

// Client represents a Supabase client with additional functionality
type Client struct {
	Client          *supabase.Client
	URL             string
	Key             string
	ServiceKey      string
	UsingServiceKey bool
}

// NewClient creates a new Supabase client
func NewClient(useServiceKey bool) (*Client, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	if url == "" || key == "" {
		return nil, errors.New("missing required environment variables: SUPABASE_URL or SUPABASE_KEY")
	}

	// Use service key if specified and available
	activeKey := key
	usingServiceKey := false
	if useServiceKey && serviceKey != "" {
		activeKey = serviceKey
		usingServiceKey = true
	}

	supabaseClient := supabase.CreateClient(url, activeKey)

	return &Client{
		Client:          supabaseClient,
		URL:             url,
		Key:             key,
		ServiceKey:      serviceKey,
		UsingServiceKey: usingServiceKey,
	}, nil
}

// Select performs a select operation on the specified table
func (c *Client) Select(table string, query string, where map[string]interface{}) ([]map[string]interface{}, error) {
	// Create request
	// Remove the second parameter which might be causing the trailing comma issue
	req := c.Client.DB.From(table).Select(query)

	// Apply where conditions if provided
	if where != nil {
		for key, value := range where {
			req.Filter(key, "eq", fmt.Sprintf("%v", value))
		}
	}

	// Execute request
	var data []map[string]interface{}
	if err := req.Execute(&data); err != nil {
		return nil, fmt.Errorf("supabase select error on %s: %v", table, err)
	}

	return data, nil
}

// Insert inserts data into the specified table
func (c *Client) Insert(table string, data interface{}) (map[string]interface{}, error) {
	// Create request
	req := c.Client.DB.From(table).Insert(data)

	// Execute request
	var result []map[string]interface{}
	if err := req.Execute(&result); err != nil {
		return nil, fmt.Errorf("supabase insert error on %s: %v", table, err)
	}

	// Return the first result if available
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, nil
}

// Update updates data in the specified table
func (c *Client) Update(table string, id string, data map[string]interface{}) (map[string]interface{}, error) {
	// Create request
	req := c.Client.DB.From(table).Update(data)

	// Apply match condition for the ID based on table name
	idColumn := "id"
	switch table {
	case "drugs":
		idColumn = "drug_id"
	case "shipments":
		idColumn = "shipment_id"
	}
	req.Filter(idColumn, "eq", id)

	// Execute request
	var result []map[string]interface{}
	if err := req.Execute(&result); err != nil {
		return nil, fmt.Errorf("supabase update error on %s: %v", table, err)
	}

	// Return the first result if available
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, nil
}

// Delete deletes data from the specified table
func (c *Client) Delete(table string, match map[string]interface{}) (map[string]interface{}, error) {
	// Create request
	req := c.Client.DB.From(table).Delete()

	// Apply match conditions
	if match != nil {
		for key, value := range match {
			req.Filter(key, "eq", fmt.Sprintf("%v", value))
		}
	}

	// Execute request
	var result map[string]interface{}
	if err := req.Execute(&result); err != nil {
		return nil, fmt.Errorf("supabase delete error on %s: %v", table, err)
	}

	return result, nil
}

// RPC calls a Postgres function with the given name and parameters
func (c *Client) RPC(functionName string, params interface{}) error {
	// Create request
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		paramsMap = map[string]interface{}{"params": params}
	}
	req := c.Client.DB.Rpc(functionName, paramsMap)

	// Execute request
	var result interface{}
	if err := req.Execute(&result); err != nil {
		return fmt.Errorf("supabase RPC error on %s: %v", functionName, err)
	}

	return nil
}

// From returns a reference to the table for building custom queries
func (c *Client) From(table string) interface{} {
	return c.Client.DB.From(table)
}
