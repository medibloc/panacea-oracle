package validation_test

import (
	"testing"

	"github.com/medibloc/panacea-oracle/validation"
	"github.com/stretchr/testify/require"
)

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

func TestValidateJSONSchema_InvalidDoc(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name"
	}`) // the required fields `description` and `body` are missing

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}

func TestValidateJSONSchema_InvalidJSON(t *testing.T) {
	schemaURI := "https://json.schemastore.org/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This JSON is messy",,,,,
	}`)

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}

func TestValidateJSONSchema_UnknownSchemaURI(t *testing.T) {
	schemaURI := "https://MED_TO_THE_MOON/github-issue-forms.json"
	jsonInput := []byte(`{
		"name": "This is a name",
		"description": "This is a description, man",
		"body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
	}`)

	err := validation.ValidateJSONSchema(jsonInput, schemaURI)
	require.Error(t, err)
}
