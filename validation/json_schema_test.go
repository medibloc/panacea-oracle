package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingleSchema(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name",
		"description": "This is a description, man",
		"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
	}`)

	schema := NewJSONSchema()
	err := schema.ValidateJSONSchema(jsonInput, schemaURI)
	require.NoError(t, err)

	require.Equal(t, 1, schema.cache.Size())
}

func TestTwoSchema(t *testing.T) {
	schemaURIs := []string{
		"https://json.schemastore.org/cdk.json",
		"https://json.schemastore.org/vsconfig.json",
	}

	schema := NewJSONSchema()
	jsonInput := []byte(`{
		"version":"1.0.0",
		"components": "",
		"versionReporting": true
	}`)
	err := schema.ValidateJSONSchemata(jsonInput, schemaURIs)
	require.ErrorContains(t, err, "components: Invalid type. Expected: array, given: string")

	require.Equal(t, 2, schema.cache.Size())
}

// TestValidateJSONSchemaInvalidDoc tests for document invalidation
func TestValidateJSONSchemaInvalidDoc(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name"
	}`) // the required fields `description` and `body` are missing

	schema := NewJSONSchema()
	err := schema.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}

// TestValidateJSONSchemaInvalidJSON tests for invalid json
func TestValidateJSONSchemaInvalidJSON(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This JSON is messy",,,,,
	}`)

	schema := NewJSONSchema()
	err := schema.ValidateJSONSchema(jsonInput, schemaURI)
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

	schema := NewJSONSchema()
	err := schema.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}
