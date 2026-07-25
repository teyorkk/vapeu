package variables

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariableResolver(t *testing.T) {
	global := map[string]string{"BASE_URL": "https://api.example.com", "ENV": "global"}
	workspace := map[string]string{"ENV": "dev", "TOKEN": "ws-token"}
	collection := map[string]string{"TOKEN": "coll-token", "VERSION": "v1"}
	request := map[string]string{"VERSION": "v2"}

	r := NewResolver(global, workspace, collection, request)

	// Scoped resolution check
	assert.Equal(t, "https://api.example.com", r.Resolve("{{BASE_URL}}"))
	assert.Equal(t, "dev", r.Resolve("{{ENV}}"))
	assert.Equal(t, "coll-token", r.Resolve("{{TOKEN}}"))
	assert.Equal(t, "v2", r.Resolve("{{VERSION}}"))

	// Combination in URL
	url := "{{BASE_URL}}/{{VERSION}}/users"
	assert.Equal(t, "https://api.example.com/v2/users", r.Resolve(url))

	// Built-in dynamic variables
	assert.NotEmpty(t, r.Resolve("{{uuid}}"))
	assert.NotEmpty(t, r.Resolve("{{timestamp}}"))
	assert.NotEmpty(t, r.Resolve("{{randomInt}}"))
	assert.NotEmpty(t, r.Resolve("{{randomString}}"))
	assert.NotEmpty(t, r.Resolve("{{hostname}}"))

	// Unresolved variable
	assert.Equal(t, "{{UNKNOWN}}", r.Resolve("{{UNKNOWN}}"))
}
