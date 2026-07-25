package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vapeu/internal/models"
	"vapeu/internal/variables"
)

type Client struct {
	httpClient *http.Client
}

type ClientOptions struct {
	TimeoutSec      int
	InsecureSSL     bool
	ProxyURL        string
	FollowRedirects bool
}

func NewClient(opts ClientOptions) *Client {
	if opts.TimeoutSec <= 0 {
		opts.TimeoutSec = 30
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.InsecureSSL,
		},
	}

	if opts.ProxyURL != "" {
		if proxyURL, err := url.Parse(opts.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	httpClient := &http.Client{
		Timeout:   time.Duration(opts.TimeoutSec) * time.Second,
		Transport: transport,
	}

	if !opts.FollowRedirects {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &Client{httpClient: httpClient}
}

func (c *Client) Execute(req *models.Request, resolver *variables.Resolver) (*models.Response, error) {
	if resolver == nil {
		resolver = variables.NewResolver(nil, nil, nil, nil)
	}

	// Resolve target URL
	resolvedURL := resolver.Resolve(req.URL)
	if !strings.HasPrefix(resolvedURL, "http://") && !strings.HasPrefix(resolvedURL, "https://") {
		resolvedURL = "http://" + resolvedURL
	}

	parsedURL, err := url.Parse(resolvedURL)
	if err != nil {
		return &models.Response{
			StatusCode: 0,
			Error:      fmt.Sprintf("Invalid URL '%s': %v", resolvedURL, err),
			Timestamp:  time.Now().Format(time.RFC3339),
		}, nil
	}

	// Query params
	q := parsedURL.Query()
	for _, qp := range req.QueryParams {
		if qp.Enabled && qp.Key != "" {
			q.Add(resolver.Resolve(qp.Key), resolver.Resolve(qp.Value))
		}
	}
	parsedURL.RawQuery = q.Encode()

	// Prepare Request Body & Content-Type
	var bodyReader io.Reader
	contentType := ""

	switch req.Body.Type {
	case models.BodyJSON:
		resolvedBody := resolver.Resolve(req.Body.Content)
		bodyReader = bytes.NewBufferString(resolvedBody)
		contentType = "application/json"
	case models.BodyXML:
		resolvedBody := resolver.Resolve(req.Body.Content)
		bodyReader = bytes.NewBufferString(resolvedBody)
		contentType = "application/xml"
	case models.BodyRaw:
		resolvedBody := resolver.Resolve(req.Body.Content)
		bodyReader = bytes.NewBufferString(resolvedBody)
		contentType = "text/plain"
	case models.BodyFormURLEncoded:
		formData := url.Values{}
		for _, item := range req.Body.FormItems {
			if item.Enabled && item.Key != "" {
				formData.Add(resolver.Resolve(item.Key), resolver.Resolve(item.Value))
			}
		}
		bodyReader = bytes.NewBufferString(formData.Encode())
		contentType = "application/x-www-form-urlencoded"
	case models.BodyMultipart:
		bodyBuf := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuf)
		for _, item := range req.Body.FormItems {
			if !item.Enabled || item.Key == "" {
				continue
			}
			key := resolver.Resolve(item.Key)
			if item.Type == "file" && item.FilePath != "" {
				filePath := resolver.Resolve(item.FilePath)
				file, err := os.Open(filePath)
				if err == nil {
					part, err := writer.CreateFormFile(key, filepath.Base(filePath))
					if err == nil {
						_, _ = io.Copy(part, file)
					}
					file.Close()
				}
			} else {
				_ = writer.WriteField(key, resolver.Resolve(item.Value))
			}
		}
		_ = writer.Close()
		bodyReader = bodyBuf
		contentType = writer.FormDataContentType()
	case models.BodyBinary:
		if req.Body.BinaryPath != "" {
			filePath := resolver.Resolve(req.Body.BinaryPath)
			fileData, err := os.ReadFile(filePath)
			if err == nil {
				bodyReader = bytes.NewReader(fileData)
				contentType = "application/octet-stream"
			}
		}
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}

	httpReq, err := http.NewRequest(method, parsedURL.String(), bodyReader)
	if err != nil {
		return &models.Response{
			StatusCode: 0,
			Error:      fmt.Sprintf("Failed to build HTTP request: %v", err),
			Timestamp:  time.Now().Format(time.RFC3339),
		}, nil
	}

	// Set Content-Type if present and not overridden
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Custom Headers
	for _, h := range req.Headers {
		if h.Enabled && h.Key != "" {
			httpReq.Header.Set(resolver.Resolve(h.Key), resolver.Resolve(h.Value))
		}
	}

	// Cookies
	for _, c := range req.Cookies {
		if c.Enabled && c.Key != "" {
			httpReq.AddCookie(&http.Cookie{
				Name:  resolver.Resolve(c.Key),
				Value: resolver.Resolve(c.Value),
			})
		}
	}

	// Auth
	switch req.Auth.Type {
	case models.AuthBasic:
		httpReq.SetBasicAuth(resolver.Resolve(req.Auth.BasicUser), resolver.Resolve(req.Auth.BasicPass))
	case models.AuthBearer:
		httpReq.Header.Set("Authorization", "Bearer "+resolver.Resolve(req.Auth.BearerToken))
	case models.AuthAPIKey:
		keyName := resolver.Resolve(req.Auth.APIKeyName)
		keyValue := resolver.Resolve(req.Auth.APIKeyValue)
		if keyName != "" {
			if req.Auth.APIKeyIn == "query" {
				q := httpReq.URL.Query()
				q.Add(keyName, keyValue)
				httpReq.URL.RawQuery = q.Encode()
			} else {
				httpReq.Header.Set(keyName, keyValue)
			}
		}
	case models.AuthCustom:
		if req.Auth.CustomHeaderName != "" {
			httpReq.Header.Set(resolver.Resolve(req.Auth.CustomHeaderName), resolver.Resolve(req.Auth.CustomHeaderValue))
		}
	}

	// Execute Request and time it
	startTime := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return &models.Response{
			StatusCode:     0,
			Error:          fmt.Sprintf("Connection error: %v", err),
			ResponseTimeMs: duration,
			Timestamp:      time.Now().Format(time.RFC3339),
		}, nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		bodyBytes = []byte{}
	}

	// Response Headers & Cookies
	respHeaders := make([]models.NameValuePair, 0, len(resp.Header))
	for k, vv := range resp.Header {
		respHeaders = append(respHeaders, models.NameValuePair{
			Key:     k,
			Value:   strings.Join(vv, ", "),
			Enabled: true,
		})
	}

	respCookies := make([]models.NameValuePair, 0, len(resp.Cookies()))
	for _, cookie := range resp.Cookies() {
		respCookies = append(respCookies, models.NameValuePair{
			Key:     cookie.Name,
			Value:   cookie.Value,
			Enabled: true,
		})
	}

	resContentType := resp.Header.Get("Content-Type")

	return &models.Response{
		StatusCode:     resp.StatusCode,
		StatusText:     resp.Status,
		Proto:          resp.Proto,
		Headers:        respHeaders,
		Cookies:        respCookies,
		Body:           bodyBytes,
		ContentType:    resContentType,
		ResponseTimeMs: duration,
		SizeBytes:      int64(len(bodyBytes)),
		Timestamp:      time.Now().Format(time.RFC3339),
	}, nil
}
