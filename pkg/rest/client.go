package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krzko/restmigrate/internal/logger"
)

type Client interface {
	SendRequest(method, endpoint string, payload interface{}) error
}

type baseClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(gatewayType, baseURL, apiKey string) (Client, error) {
	base := &baseClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}

	switch gatewayType {
	case "apisix":
		return &APISIXClient{baseClient: base}, nil
	case "kong":
		return &KongClient{baseClient: base}, nil
	case "generic":
		return &GenericClient{baseClient: base}, nil
	default:
		return nil, fmt.Errorf("unsupported gateway type: %s", gatewayType)
	}
}

func (c *baseClient) sendRequest(method, endpoint string, payload interface{}, headers map[string]string) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	logger.Debug("Sending request", "method", method, "url", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	logger.Debug("Request successful", "method", method, "url", url, "status", resp.Status)
	return nil
}

type APISIXClient struct {
	*baseClient
}

func (c *APISIXClient) SendRequest(method, endpoint string, payload interface{}) error {
	headers := map[string]string{
		"X-API-KEY": c.apiKey,
	}
	return c.sendRequest(method, endpoint, payload, headers)
}

type KongClient struct {
	*baseClient
}

func (c *KongClient) SendRequest(method, endpoint string, payload interface{}) error {
	headers := map[string]string{
		"Kong-Admin-Token": c.apiKey,
	}
	return c.sendRequest(method, endpoint, payload, headers)
}

type GenericClient struct {
	*baseClient
}

func (c *GenericClient) SendRequest(method, endpoint string, payload interface{}) error {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.apiKey),
	}
	return c.sendRequest(method, endpoint, payload, headers)
}
