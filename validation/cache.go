package validation

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

type schemaCache struct {
	cache map[string]*gojsonschema.Schema
}

func newSchemaCache() *schemaCache {
	return &schemaCache{make(map[string]*gojsonschema.Schema)}
}

// Get returns a JSON Schema of the passed URI.
// It calls network once first time, and stores it in memory for later use.
func (l *schemaCache) Get(schemaURI string) (*gojsonschema.Schema, error) {
	if schema, ok := l.cache[schemaURI]; ok {
		return schema, nil
	}

	schema, err := newReferenceSchema(schemaURI)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema. %w", err)
	}
	l.cache[schemaURI] = schema

	return schema, nil
}

// Size returns the size of the current cache.
func (l *schemaCache) Size() int {
	return len(l.cache)
}
