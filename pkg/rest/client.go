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
	"go.opentelemetry.io/otel/trace"
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

	req, err := c.createRequest(ctx, method, url, payload, headers)
	if err != nil {
		return c.handleError(span, "Failed to create request", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.handleError(span, "Failed to send request", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.handleError(span, "Failed to read response body", err)
	}

	c.setSpanAttributes(span, method, url, resp.StatusCode, string(responseBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleErrorResponse(span, method, url, resp.StatusCode, string(responseBody))
	}

	c.logSuccess(method, url, resp.Status, string(responseBody))
	span.SetStatus(codes.Ok, "")
	return nil
}

func (c *baseClient) createRequest(ctx context.Context, method, url string, payload interface{}, headers map[string]string) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	return req, nil
}

func (c *baseClient) handleError(span trace.Span, msg string, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, msg)
	return fmt.Errorf("%s: %w", msg, err)
}

func (c *baseClient) handleErrorResponse(span trace.Span, method, url string, statusCode int, responseBody string) error {
	errorResp := &ErrorResponse{
		StatusCode: statusCode,
		Body:       responseBody,
	}
	span.RecordError(errorResp)
	span.SetStatus(codes.Error, fmt.Sprintf("Request failed with status code %d", statusCode))
	logger.Error("Request failed",
		"method", method,
		"url", url,
		"status", statusCode,
		"response", responseBody)
	return errorResp
}

func (c *baseClient) setSpanAttributes(span trace.Span, method, url string, statusCode int, responseBody string) {
	span.SetAttributes(
		attribute.Int("http.status_code", statusCode),
		attribute.String("http.method", method),
		attribute.String("http.url", url),
		attribute.String("http.response_body", responseBody),
	)
}

func (c *baseClient) logSuccess(method, url, status, responseBody string) {
	logger.Debug("Request successful",
		"method", method,
		"url", url,
		"status", status,
		"response", responseBody)
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
