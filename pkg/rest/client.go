package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krzko/restmigrate/internal/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

type Client interface {
	SendRequest(ctx context.Context, method, endpoint string, payload interface{}) error
}

type baseClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

func NewClient(gatewayType, baseURL, apiKey string) (Client, error) {
	base := &baseClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
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

func (c *baseClient) sendRequest(ctx context.Context, method, endpoint string, payload interface{}, headers map[string]string) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	ctx, span := otel.Tracer("restmigrate/client").Start(ctx, fmt.Sprintf("%s %s", method, endpoint))
	defer span.End()

	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to marshal payload")
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Inject traceparent header
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	logger.Debug("Sending request", "method", method, "url", url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to send request")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read response body")
		return fmt.Errorf("failed to read response body: %w", err)
	}

	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("http.method", method),
		attribute.String("http.url", url),
		attribute.String("http.response_body", string(responseBody)),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorResp := &ErrorResponse{
			StatusCode: resp.StatusCode,
			Body:       string(responseBody),
		}
		span.RecordError(errorResp)
		span.SetStatus(codes.Error, fmt.Sprintf("Request failed with status code %d", resp.StatusCode))
		logger.Error("Request failed",
			"method", method,
			"url", url,
			"status", resp.StatusCode,
			"response", string(responseBody))
		return errorResp
	}

	span.SetStatus(codes.Ok, "")
	logger.Debug("Request successful",
		"method", method,
		"url", url,
		"status", resp.Status,
		"response", string(responseBody))
	return nil
}

type APISIXClient struct {
	*baseClient
}

func (c *APISIXClient) SendRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	headers := map[string]string{
		"X-API-KEY": c.apiKey,
	}
	return c.sendRequest(ctx, method, endpoint, payload, headers)
}

type KongClient struct {
	*baseClient
}

func (c *KongClient) SendRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	headers := map[string]string{
		"Kong-Admin-Token": c.apiKey,
	}
	return c.sendRequest(ctx, method, endpoint, payload, headers)
}

type GenericClient struct {
	*baseClient
}

func (c *GenericClient) SendRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.apiKey),
	}
	return c.sendRequest(ctx, method, endpoint, payload, headers)
}
