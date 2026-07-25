package impexp

import (
	"testing"

	"vapeu/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestParseCurl(t *testing.T) {
	curlCmd := `curl -X POST "https://api.example.com/login" -H "Content-Type: application/json" -H "Authorization: Bearer testtoken123" -d '{"username":"admin"}'`

	req, err := ParseCurl(curlCmd)
	assert.NoError(t, err)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://api.example.com/login", req.URL)
	assert.Equal(t, models.AuthBearer, req.Auth.Type)
	assert.Equal(t, "testtoken123", req.Auth.BearerToken)
	assert.Equal(t, models.BodyJSON, req.Body.Type)
	assert.Equal(t, `{"username":"admin"}`, req.Body.Content)

	// Convert back to cURL
	curlOutput := RequestToCurl(req)
	assert.Contains(t, curlOutput, "curl -X POST")
	assert.Contains(t, curlOutput, "https://api.example.com/login")
	assert.Contains(t, curlOutput, "Authorization: Bearer testtoken123")
}

func TestImportOpenAPI(t *testing.T) {
	openapiYAML := `
openapi: 3.0.0
info:
  title: Test API
paths:
  /users:
    get:
      summary: List users
    post:
      summary: Create user
`
	col, err := ImportOpenAPI([]byte(openapiYAML))
	assert.NoError(t, err)
	assert.Equal(t, "Test API", col.Name)
	assert.Len(t, col.Nodes, 2)
}
