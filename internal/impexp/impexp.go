package impexp

import (
	"encoding/json"
	"fmt"
	"strings"

	"vapeu/internal/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// ParseCurl parses a cURL command into a models.Request struct
func ParseCurl(curlCmd string) (*models.Request, error) {
	tokens := tokenizeCommand(curlCmd)
	if len(tokens) == 0 || (tokens[0] != "curl" && tokens[0] != "curl.exe") {
		return nil, fmt.Errorf("invalid cURL command: must start with 'curl'")
	}

	req := &models.Request{
		ID:     uuid.New().String(),
		Name:   "Imported Request",
		Method: "GET",
		Body:   models.RequestBody{Type: models.BodyNone},
		Auth:   models.AuthConfig{Type: models.AuthNone},
	}

	var headers []models.NameValuePair
	var bodyStr string
	hasExplicitMethod := false

	i := 1
	for i < len(tokens) {
		token := tokens[i]
		switch token {
		case "-X", "--request":
			if i+1 < len(tokens) {
				req.Method = strings.ToUpper(tokens[i+1])
				hasExplicitMethod = true
				i += 2
				continue
			}
		case "-H", "--header":
			if i+1 < len(tokens) {
				headerVal := tokens[i+1]
				parts := strings.SplitN(headerVal, ":", 2)
				if len(parts) == 2 {
					k := strings.TrimSpace(parts[0])
					v := strings.TrimSpace(parts[1])

					if strings.ToLower(k) == "authorization" && strings.HasPrefix(strings.ToLower(v), "bearer ") {
						req.Auth = models.AuthConfig{
							Type:        models.AuthBearer,
							BearerToken: strings.TrimSpace(v[7:]),
						}
					} else {
						headers = append(headers, models.NameValuePair{Key: k, Value: v, Enabled: true})
					}
				}
				i += 2
				continue
			}
		case "-d", "--data", "--data-raw", "--data-binary":
			if i+1 < len(tokens) {
				bodyStr = tokens[i+1]
				if !hasExplicitMethod {
					req.Method = "POST"
				}
				i += 2
				continue
			}
		case "-u", "--user":
			if i+1 < len(tokens) {
				userPass := tokens[i+1]
				parts := strings.SplitN(userPass, ":", 2)
				pass := ""
				if len(parts) > 1 {
					pass = parts[1]
				}
				req.Auth = models.AuthConfig{
					Type:      models.AuthBasic,
					BasicUser: parts[0],
					BasicPass: pass,
				}
				i += 2
				continue
			}
		case "-L", "--location":
			i++
			continue
		case "-k", "--insecure":
			i++
			continue
		default:
			if strings.HasPrefix(token, "-") {
				i++
				continue
			}
			// Potential URL
			if req.URL == "" && (strings.HasPrefix(token, "http://") || strings.HasPrefix(token, "https://") || strings.Contains(token, ".")) {
				req.URL = token
			}
			i++
		}
	}

	req.Headers = headers

	if bodyStr != "" {
		req.Body.Content = bodyStr
		isJSON := false
		for _, h := range headers {
			if strings.ToLower(h.Key) == "content-type" && strings.Contains(strings.ToLower(h.Value), "json") {
				isJSON = true
				break
			}
		}
		if isJSON || (strings.HasPrefix(strings.TrimSpace(bodyStr), "{") || strings.HasPrefix(strings.TrimSpace(bodyStr), "[")) {
			req.Body.Type = models.BodyJSON
		} else {
			req.Body.Type = models.BodyRaw
		}
	}

	if req.URL == "" {
		req.URL = "https://httpbin.org/get"
	}

	return req, nil
}

// RequestToCurl converts a models.Request into a runnable cURL string
func RequestToCurl(req *models.Request) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("curl -X %s \"%s\"", req.Method, req.URL))

	for _, h := range req.Headers {
		if h.Enabled && h.Key != "" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H \"%s: %s\"", h.Key, h.Value))
		}
	}

	// Auth headers
	switch req.Auth.Type {
	case models.AuthBearer:
		if req.Auth.BearerToken != "" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H \"Authorization: Bearer %s\"", req.Auth.BearerToken))
		}
	case models.AuthBasic:
		sb.WriteString(fmt.Sprintf(" \\\n  -u \"%s:%s\"", req.Auth.BasicUser, req.Auth.BasicPass))
	case models.AuthAPIKey:
		if req.Auth.APIKeyIn == "header" && req.Auth.APIKeyName != "" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H \"%s: %s\"", req.Auth.APIKeyName, req.Auth.APIKeyValue))
		}
	case models.AuthCustom:
		if req.Auth.CustomHeaderName != "" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H \"%s: %s\"", req.Auth.CustomHeaderName, req.Auth.CustomHeaderValue))
		}
	}

	if req.Body.Type != models.BodyNone && req.Body.Content != "" {
		// Escape single quotes for shell
		escapedBody := strings.ReplaceAll(req.Body.Content, "'", "'\\''")
		sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", escapedBody))
	}

	return sb.String()
}

func ImportOpenAPI(data []byte) (*models.Collection, error) {
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		if errJson := json.Unmarshal(data, &doc); errJson != nil {
			return nil, fmt.Errorf("failed to parse OpenAPI doc: %w", err)
		}
	}

	info, _ := doc["info"].(map[string]interface{})
	title := "OpenAPI Collection"
	if info != nil {
		if t, ok := info["title"].(string); ok && t != "" {
			title = t
		}
	}

	col := &models.Collection{
		ID:        uuid.New().String(),
		Name:      title,
		Nodes:     []*models.CollectionNode{},
		Variables: make(map[string]string),
	}

	paths, ok := doc["paths"].(map[string]interface{})
	if !ok {
		return col, nil
	}

	for pathStr, methods := range paths {
		methodMap, ok := methods.(map[string]interface{})
		if !ok {
			continue
		}

		for method, op := range methodMap {
			mUpper := strings.ToUpper(method)
			if mUpper != "GET" && mUpper != "POST" && mUpper != "PUT" && mUpper != "DELETE" && mUpper != "PATCH" && mUpper != "HEAD" && mUpper != "OPTIONS" {
				continue
			}

			opMap, _ := op.(map[string]interface{})
			summary := pathStr
			if opMap != nil {
				if s, ok := opMap["summary"].(string); ok && s != "" {
					summary = s
				}
			}

			req := &models.Request{
				ID:          uuid.New().String(),
				Name:        fmt.Sprintf("%s %s", mUpper, summary),
				Method:      mUpper,
				URL:         "{{BASE_URL}}" + pathStr,
				Headers:     []models.NameValuePair{},
				QueryParams: []models.NameValuePair{},
				Auth:        models.AuthConfig{Type: models.AuthNone},
				Body:        models.RequestBody{Type: models.BodyNone},
			}

			col.Nodes = append(col.Nodes, &models.CollectionNode{
				ID:      uuid.New().String(),
				Name:    req.Name,
				Kind:    models.NodeRequest,
				Request: req,
			})
		}
	}

	return col, nil
}

func ImportPostmanCollection(data []byte) (*models.Collection, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	info, _ := doc["info"].(map[string]interface{})
	name := "Postman Collection"
	if info != nil {
		if n, ok := info["name"].(string); ok {
			name = n
		}
	}

	col := &models.Collection{
		ID:        uuid.New().String(),
		Name:      name,
		Nodes:     []*models.CollectionNode{},
		Variables: make(map[string]string),
	}

	items, _ := doc["item"].([]interface{})
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		node := parsePostmanItem(itemMap)
		if node != nil {
			col.Nodes = append(col.Nodes, node)
		}
	}

	return col, nil
}

func parsePostmanItem(item map[string]interface{}) *models.CollectionNode {
	name, _ := item["name"].(string)
	if name == "" {
		name = "Untitled"
	}

	// Check if folder (has item array)
	if subItems, ok := item["item"].([]interface{}); ok {
		node := &models.CollectionNode{
			ID:       uuid.New().String(),
			Name:     name,
			Kind:     models.NodeFolder,
			Children: []*models.CollectionNode{},
		}
		for _, sub := range subItems {
			if subMap, ok := sub.(map[string]interface{}); ok {
				if childNode := parsePostmanItem(subMap); childNode != nil {
					node.Children = append(node.Children, childNode)
				}
			}
		}
		return node
	}

	// Request item
	reqData, ok := item["request"].(map[string]interface{})
	if !ok {
		return nil
	}

	method, _ := reqData["method"].(string)
	if method == "" {
		method = "GET"
	}

	urlStr := ""
	if urlMap, ok := reqData["url"].(map[string]interface{}); ok {
		if raw, ok := urlMap["raw"].(string); ok {
			urlStr = raw
		}
	} else if u, ok := reqData["url"].(string); ok {
		urlStr = u
	}

	req := &models.Request{
		ID:          uuid.New().String(),
		Name:        name,
		Method:      strings.ToUpper(method),
		URL:         urlStr,
		Headers:     []models.NameValuePair{},
		QueryParams: []models.NameValuePair{},
		Auth:        models.AuthConfig{Type: models.AuthNone},
		Body:        models.RequestBody{Type: models.BodyNone},
	}

	return &models.CollectionNode{
		ID:      uuid.New().String(),
		Name:    name,
		Kind:    models.NodeRequest,
		Request: req,
	}
}

func tokenizeCommand(cmd string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(cmd); i++ {
		r := cmd[i]

		if escaped {
			current.WriteByte(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
			escaped = true
			continue
		}

		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inSingle && !inDouble {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(r)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
