package datadeal

import (
	"crypto/sha256"
	"testing"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/stretchr/testify/require"
)

func TestCanonicalJSON(t *testing.T) {
	jsonDataBz := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// WhiteSpace before "name"
	jsonDataBzSpace := []byte(
		`
		{
			"invalid_key_name":  "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// WhiteSpace before "}" bracket
	jsonDataBzBracket := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		 }
		`)

	jsonDataBzOrder := []byte(
		`
		{
			"invalid_key_description": "description",
			"invalid_key_name": "name",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	jsonData, err := jsoncanonicalizer.Transform(jsonDataBz)
	require.NoError(t, err)
	dataHash := sha256.Sum256(jsonData)

	jsonDataSpace, err := jsoncanonicalizer.Transform(jsonDataBzSpace)
	require.NoError(t, err)
	dataHashSpace := sha256.Sum256(jsonDataSpace)

	jsonDataOrder, err := jsoncanonicalizer.Transform(jsonDataBzOrder)
	require.NoError(t, err)
	dataHashOrder := sha256.Sum256(jsonDataOrder)

	jsonDataBracket, err := jsoncanonicalizer.Transform(jsonDataBzBracket)
	require.NoError(t, err)
	dataHashBracket := sha256.Sum256(jsonDataBracket)

	require.Equal(t, dataHash, dataHashOrder)
	require.Equal(t, dataHash, dataHashSpace)
	require.Equal(t, dataHashOrder, dataHashSpace)
	require.Equal(t, dataHash, dataHashBracket)
}

func TestCanonicalJSON_DataHash_Not_Equal(t *testing.T) {
	jsonDataBz := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// WhiteSpace before "name"
	jsonDataBzSpace := []byte(
		`
		{
			"invalid_key_name":  "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	// WhiteSpace before "}" bracket
	jsonDataBzBracket := []byte(
		`
		{
			"invalid_key_name": "name",
			"invalid_key_description": "description",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		 }
		`)

	jsonDataBzOrder := []byte(
		`
		{
			"invalid_key_description": "description",
			"invalid_key_name": "name",
			"invalid_key_body": [{ "type": "markdown", "attributes": { "value": "val1" } }]
		}
		`)

	jsonData, err := jsoncanonicalizer.Transform(jsonDataBz)
	require.NoError(t, err)
	dataHash := sha256.Sum256(jsonData)

	dataHashSpace := sha256.Sum256(jsonDataBzSpace)
	dataHashBracket := sha256.Sum256(jsonDataBzBracket)
	dataHashOrder := sha256.Sum256(jsonDataBzOrder)

	require.NotEqual(t, dataHash, dataHashSpace)
	require.NotEqual(t, dataHash, dataHashBracket)
	require.NotEqual(t, dataHash, dataHashOrder)
}
