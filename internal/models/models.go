package models

import (
	"time"
)

type NameValuePair struct {
	Key         string `json:"key" yaml:"key"`
	Value       string `json:"value" yaml:"value"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBasic  AuthType = "basic"
	AuthBearer AuthType = "bearer"
	AuthAPIKey AuthType = "apikey"
	AuthCustom AuthType = "custom"
)

type AuthConfig struct {
	Type              AuthType `json:"type" yaml:"type"`
	BasicUser         string   `json:"basic_user,omitempty" yaml:"basic_user,omitempty"`
	BasicPass         string   `json:"basic_pass,omitempty" yaml:"basic_pass,omitempty"`
	BearerToken       string   `json:"bearer_token,omitempty" yaml:"bearer_token,omitempty"`
	APIKeyName        string   `json:"api_key_name,omitempty" yaml:"api_key_name,omitempty"`
	APIKeyValue        string   `json:"api_key_value,omitempty" yaml:"api_key_value,omitempty"`
	APIKeyIn          string   `json:"api_key_in,omitempty" yaml:"api_key_in,omitempty"` // "header" or "query"
	CustomHeaderName  string   `json:"custom_header_name,omitempty" yaml:"custom_header_name,omitempty"`
	CustomHeaderValue string   `json:"custom_header_value,omitempty" yaml:"custom_header_value,omitempty"`
}

type BodyType string

const (
	BodyNone           BodyType = "none"
	BodyJSON           BodyType = "json"
	BodyRaw            BodyType = "raw"
	BodyXML            BodyType = "xml"
	BodyFormURLEncoded BodyType = "form_urlencoded"
	BodyMultipart      BodyType = "multipart"
	BodyBinary         BodyType = "binary"
)

type FormDataItem struct {
	Key      string `json:"key" yaml:"key"`
	Value    string `json:"value" yaml:"value"`
	Type     string `json:"type" yaml:"type"` // "text" or "file"
	FilePath string `json:"file_path,omitempty" yaml:"file_path,omitempty"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

type RequestBody struct {
	Type       BodyType       `json:"type" yaml:"type"`
	Content    string         `json:"content,omitempty" yaml:"content,omitempty"`
	FormItems  []FormDataItem `json:"form_items,omitempty" yaml:"form_items,omitempty"`
	BinaryPath string         `json:"binary_path,omitempty" yaml:"binary_path,omitempty"`
}

type Request struct {
	ID          string          `json:"id" yaml:"id"`
	Name        string          `json:"name" yaml:"name"`
	Method      string          `json:"method" yaml:"method"`
	URL         string          `json:"url" yaml:"url"`
	Headers     []NameValuePair `json:"headers" yaml:"headers"`
	QueryParams []NameValuePair `json:"query_params" yaml:"query_params"`
	Cookies     []NameValuePair `json:"cookies" yaml:"cookies"`
	Auth        AuthConfig      `json:"auth" yaml:"auth"`
	Body        RequestBody     `json:"body" yaml:"body"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type Response struct {
	StatusCode     int             `json:"status_code" yaml:"status_code"`
	StatusText     string          `json:"status_text" yaml:"status_text"`
	Proto          string          `json:"proto" yaml:"proto"`
	Headers        []NameValuePair `json:"headers" yaml:"headers"`
	Cookies        []NameValuePair `json:"cookies" yaml:"cookies"`
	Body           []byte          `json:"body" yaml:"body"`
	ContentType    string          `json:"content_type" yaml:"content_type"`
	ResponseTimeMs int64           `json:"response_time_ms" yaml:"response_time_ms"`
	SizeBytes      int64           `json:"size_bytes" yaml:"size_bytes"`
	Error          string          `json:"error,omitempty" yaml:"error,omitempty"`
	Timestamp      string          `json:"timestamp" yaml:"timestamp"`
}

type CollectionNodeKind string

const (
	NodeFolder  CollectionNodeKind = "folder"
	NodeRequest CollectionNodeKind = "request"
)

type CollectionNode struct {
	ID          string             `json:"id" yaml:"id"`
	Name        string             `json:"name" yaml:"name"`
	Kind        CollectionNodeKind `json:"kind" yaml:"kind"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Request     *Request           `json:"request,omitempty" yaml:"request,omitempty"`
	Children    []*CollectionNode  `json:"children,omitempty" yaml:"children,omitempty"`
}

type Collection struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Nodes       []*CollectionNode `json:"nodes" yaml:"nodes"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type Environment struct {
	ID        string            `json:"id" yaml:"id"`
	Name      string            `json:"name" yaml:"name"`
	Variables map[string]string `json:"variables" yaml:"variables"`
	IsGlobal  bool              `json:"is_global" yaml:"is_global"`
}

type HistoryItem struct {
	ID         string    `json:"id" yaml:"id"`
	Request    Request   `json:"request" yaml:"request"`
	Response   *Response `json:"response,omitempty" yaml:"response,omitempty"`
	Timestamp  time.Time `json:"timestamp" yaml:"timestamp"`
	DurationMs int64     `json:"duration_ms" yaml:"duration_ms"`
	StatusCode int       `json:"status_code" yaml:"status_code"`
	IsFavorite bool      `json:"is_favorite" yaml:"is_favorite"`
}

type Workspace struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	ActiveEnvID string   `json:"active_env_id,omitempty" yaml:"active_env_id,omitempty"`
	Collections []string `json:"collections,omitempty" yaml:"collections,omitempty"`
}

type Config struct {
	Theme             string          `json:"theme" yaml:"theme"`
	Editor            string          `json:"editor" yaml:"editor"`
	DefaultTimeoutSec int             `json:"default_timeout_sec" yaml:"default_timeout_sec"`
	SSLVerification   bool            `json:"ssl_verification" yaml:"ssl_verification"`
	ProxyURL          string          `json:"proxy_url,omitempty" yaml:"proxy_url,omitempty"`
	DefaultHeaders    []NameValuePair `json:"default_headers,omitempty" yaml:"default_headers,omitempty"`
}

func NewDefaultConfig() Config {
	return Config{
		Theme:             "dark",
		Editor:            "vim",
		DefaultTimeoutSec: 30,
		SSLVerification:   true,
		ProxyURL:          "",
		DefaultHeaders: []NameValuePair{
			{Key: "User-Agent", Value: "apicli/1.0", Enabled: true},
		},
	}
}
