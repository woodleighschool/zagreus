package nessus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	tokenRefreshLeeway = 30 * time.Second
)

type Config struct {
	Host      string
	AccessKey string
	SecretKey string
}

type Client struct {
	apiEndpoint string
	accessKey   string
	secretKey   string
	httpClient  *http.Client
}

func NewClient(config Config) (*Client, error) {
	if strings.TrimSpace(config.AccessKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		return nil, fmt.Errorf("nessus Access Key and Secret Key are required")
	}
	parsed, err := url.Parse(config.Host)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, fmt.Errorf("a Nessus Professional host is required: https://{your_host}")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("nessus host must not contain query or fragment")
	}
	apiEndpoint := strings.TrimRight(config.Host, "/")
	return &Client{
		apiEndpoint: apiEndpoint,
		accessKey:   config.AccessKey,
		secretKey:   config.SecretKey,
		httpClient:  &http.Client{Timeout: defaultHTTPTimeout},
	}, nil
}

func (c *Client) request(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	payload any,
	successCodes ...int,
) ([]byte, error) {
	encoded, err := encodePayload(payload)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s%s", c.apiEndpoint, path)
	if len(query) != 0 {
		requestURL += "?" + query.Encode()
	}
	request, requestErr := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(encoded))
	if requestErr != nil {
		return nil, fmt.Errorf("create nessus api request: %w", requestErr)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Apikeys", fmt.Sprintf("accessKey=%s; secretKey=%s", c.accessKey, c.secretKey))

	response, requestErr := c.httpClient.Do(request)
	if requestErr != nil {
		return nil, fmt.Errorf("call nessus api: %w", requestErr)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read nessus response: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close nessus response: %w", closeErr)
	}
	if !containsStatus(successCodes, response.StatusCode) {
		return nil, NewHTTPError(requestURL, response.StatusCode, body)
	}
	return body, nil
}

func encodePayload(payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode nessus request body: %w", err)
	}
	return encoded, nil
}

func containsStatus(statuses []int, status int) bool {
	return slices.Contains(statuses, status)
}
