package validation_test

import (
	"testing"

	"github.com/medibloc/panacea-oracle/validation"
	"github.com/stretchr/testify/require"
)

// TestValidateJSONSchema tests for JSONSchema validation success
func TestValidateJSONSchema(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name",
		"description": "This is a description, man",
		"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
	}`)

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.NoError(t, err)
}

// TestValidateJSONSchemaInvalidDoc tests for document invalidation
func TestValidateJSONSchemaInvalidDoc(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name"
	}`) // the required fields `description` and `body` are missing

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}

// TestValidateJSONSchemaInvalidJSON tests for invalid json
func TestValidateJSONSchemaInvalidJSON(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This JSON is messy",,,,,
	}`)

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}

// TestValidateJSONSchemaUnknownSchemaURI tests for invalid schemaURIs.
func TestValidateJSONSchemaUnknownSchemaURI(t *testing.T) {
	schemaURI := "https://MED_TO_THE_MOON/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name",
		"description": "This is a description, man",
		"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
	}`)

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}
