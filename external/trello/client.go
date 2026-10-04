package trello

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
	"sync"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	tokenRefreshLeeway = 30 * time.Second
)

type Config struct {
	AccessKey   string
	AccessToken string
}

type Client struct {
	apiEndpoint string
	accessKey   string
	accessToken string
	httpClient  *http.Client
	now         func() time.Time

	tokenMu sync.Mutex
	token   string
}

func NewClient(config Config) (*Client, error) {
	if strings.TrimSpace(config.AccessKey) == "" || strings.TrimSpace(config.AccessToken) == "" {
		return nil, fmt.Errorf("Trello Access Key and Secret Key are required")
	}
	return &Client{
		apiEndpoint: "https://api.trello.com/1",
		accessKey:   config.AccessKey,
		accessToken: config.AccessToken,
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
	requestURL := fmt.Sprintf("%s/%s", c.apiEndpoint, path)
	if len(query) != 0 {
		requestURL += "?" + query.Encode()
	}
	request, requestErr := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(encoded))
	if requestErr != nil {
		return nil, fmt.Errorf("create trello api request: %w", requestErr)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("OAuth oauth_consumer_key=\"%s\", oauth_token=\"%s\"", c.accessKey, c.accessToken))

	response, requestErr := c.httpClient.Do(request)
	if requestErr != nil {
		return nil, fmt.Errorf("call trello api: %w", requestErr)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read trello response: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close trello response: %w", closeErr)
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
		return nil, fmt.Errorf("encode trello request body: %w", err)
	}
	return encoded, nil
}

func containsStatus(statuses []int, status int) bool {
	return slices.Contains(statuses, status)
}
