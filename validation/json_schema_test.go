package validation

import (
	"github.com/stretchr/testify/require"
	"testing"
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
		"https://oracle-testnet.s3.ap-northeast-2.amazonaws.com/performance_schema.json",
		"https://json.schemastore.org/cdk.json",
	}

	schema := NewJSONSchema()
	jsonInput := []byte(`{
		"id": "This is a ID",
		"name": "This is a name",
		"age": 39,
		"title": "This is a title",
		"description": "This is a description",
		"versionReporting": "Reporting!"
	}`)
	err := schema.ValidateJSONSchemata(jsonInput, schemaURIs)
	require.ErrorContains(t, err, "Invalid type. Expected: boolean, given: string")

	require.Equal(t, 2, schema.cache.Size())
}
