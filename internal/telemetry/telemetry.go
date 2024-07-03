package telemetry

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/krzko/restmigrate/internal/logger"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

const (
	OtelEndpointEnvVar    = "OTEL_EXPORTER_OTLP_ENDPOINT"
	OtelInsecureEnvVar    = "OTEL_EXPORTER_OTLP_INSECURE"
	OtelSDKDisabledEnvVar = "OTEL_SDK_DISABLED"
)

var tracer trace.Tracer

type exporterResult struct {
	exporter *otlptrace.Exporter
	err      error
}

type OtelConfig struct {
	endpoint    string
	insecure    bool
	sdkDisabled bool
}

func parseOtelConfig() OtelConfig {
	config := OtelConfig{}
	config.endpoint = os.Getenv(OtelEndpointEnvVar)
	if config.endpoint == "" {
		config.endpoint = "localhost:4317" // Default endpoint
	}
	config.insecure, _ = strconv.ParseBool(os.Getenv(OtelInsecureEnvVar))

	// Telemetry disabled by default
	sdkEnabled, _ := strconv.ParseBool(os.Getenv("OTEL_SDK_ENABLED"))
	config.sdkDisabled = !sdkEnabled

	return config
}

func getTraceExporter(ctx context.Context, config OtelConfig) (*otlptrace.Exporter, error) {
	logger.Debug("Initialising trace exporter", "endpoint", config.endpoint, "insecure", config.insecure)

	var secureOption otlptracegrpc.Option
	if config.insecure {
		secureOption = otlptracegrpc.WithInsecure()
	} else {
		secureOption = otlptracegrpc.WithTLSCredentials(nil)
	}

	retryConfig := otlptracegrpc.RetryConfig{
		Enabled:         true,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		MaxElapsedTime:  30 * time.Second,
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.endpoint),
		secureOption,
		otlptracegrpc.WithDialOption(grpc.WithDisableServiceConfig()),
		otlptracegrpc.WithRetry(retryConfig),
		otlptracegrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	logger.Debug("Trace exporter initialised successfully")
	return exporter, nil
}

func InitTracer(serviceName string, attributes map[string]string) (func(context.Context) error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := parseOtelConfig()

	if config.sdkDisabled {
		return setupNoopTracer(serviceName)
	}

	return setupActiveTracer(ctx, serviceName, attributes, config)
}

func setupNoopTracer(serviceName string) (func(context.Context) error, error) {
	logger.Info("OpenTelemetry", "status", "disabled")
	tracer = trace.NewNoopTracerProvider().Tracer(serviceName)
	return func(context.Context) error { return nil }, nil
}

func setupActiveTracer(ctx context.Context, serviceName string, attributes map[string]string, config OtelConfig) (func(context.Context) error, error) {
	logger.Info("Initialising OpenTelemetry", "endpoint", config.endpoint)

	res, err := setupResource(attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to setup resource: %w", err)
	}

	exporter, err := setupExporter(ctx, config)
	if err != nil {
		return nil, err
	}

	tracerProvider := setupTracerProvider(res, exporter)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	tracer = tracerProvider.Tracer(serviceName)

	logger.Info("OpenTelemetry", "status", "enabled")

	return createShutdownFunction(tracerProvider, exporter), nil
}

func setupResource(attributes map[string]string) (*resource.Resource, error) {
	res, err := DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to detect environment: %w", err)
	}

	for k, v := range attributes {
		res, _ = resource.Merge(res, resource.NewWithAttributes(semconv.SchemaURL, attribute.String(k, v)))
	}

	return res, nil
}

func setupExporter(ctx context.Context, config OtelConfig) (*otlptrace.Exporter, error) {
	exporterChan := make(chan exporterResult)
	go func() {
		exporter, err := getTraceExporter(ctx, config)
		exporterChan <- exporterResult{exporter, err}
	}()

	select {
	case result := <-exporterChan:
		if result.err != nil {
			logger.Error("Failed to initialise exporter", "error", result.err)
			return nil, result.err
		}
		return result.exporter, nil
	case <-time.After(20 * time.Second):
		logger.Warn("Exporter initialization timed out, continuing with a noop exporter")
		return otlptrace.NewUnstarted(nil), nil
	}
}

func setupTracerProvider(res *resource.Resource, exporter *otlptrace.Exporter) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
	)
}

func createShutdownFunction(tracerProvider *sdktrace.TracerProvider, exporter *otlptrace.Exporter) func(context.Context) error {
	return func(ctx context.Context) error {
		logger.Debug("Shutting down OpenTelemetry")

		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down tracer provider", "error", err)
		} else {
			logger.Debug("Tracer provider shut down successfully")
		}

		if exporter != nil {
			if err := exporter.Shutdown(shutdownCtx); err != nil {
				logger.Error("Error shutting down exporter", "error", err)
			} else {
				logger.Debug("Exporter shut down successfully")
			}
		}

		logger.Debug("OpenTelemetry shut down process completed")
		return nil
	}
}

func SetSpanStatus(span trace.Span, err error, attrs ...attribute.KeyValue) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attrs...)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}
