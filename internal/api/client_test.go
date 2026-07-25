package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vapeu/internal/models"
	"vapeu/internal/variables"
	"github.com/stretchr/testify/assert"
)

func TestClientExecuteJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer mysecrettoken", r.Header.Get("Authorization"))

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "foo_val", body["foo"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := NewClient(ClientOptions{TimeoutSec: 5})

	req := &models.Request{
		Method: "POST",
		URL:    ts.URL,
		Auth: models.AuthConfig{
			Type:        models.AuthBearer,
			BearerToken: "{{TOKEN}}",
		},
		Body: models.RequestBody{
			Type:    models.BodyJSON,
			Content: `{"foo": "{{FOO}}"}`,
		},
	}

	resolver := variables.NewResolver(
		map[string]string{"TOKEN": "mysecrettoken", "FOO": "foo_val"},
		nil, nil, nil,
	)

	resp, err := client.Execute(req, resolver)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Empty(t, resp.Error)
	assert.Contains(t, string(resp.Body), `"status":"ok"`)
	assert.True(t, resp.ResponseTimeMs >= 0)
}
