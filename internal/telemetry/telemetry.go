package telemetry

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/krzko/restmigrate/internal/logger"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	OtelEndpointEnvVar    = "OTEL_EXPORTER_OTLP_ENDPOINT"
	OtelInsecureEnvVar    = "OTEL_EXPORTER_OTLP_INSECURE"
	OtelSDKDisabledEnvVar = "OTEL_SDK_DISABLED"
)

var tracer trace.Tracer

type OtelConfig struct {
	endpoint    string
	insecure    bool
	sdkDisabled bool
}

func parseOtelConfig() OtelConfig {
	config := OtelConfig{}
	config.endpoint = os.Getenv(OtelEndpointEnvVar)
	insecure, err := strconv.ParseBool(os.Getenv(OtelInsecureEnvVar))
	if err != nil {
		insecure = true
	}
	config.insecure = insecure

	sdkDisabled, err := strconv.ParseBool(os.Getenv(OtelSDKDisabledEnvVar))
	if err != nil {
		sdkDisabled = false
	}
	config.sdkDisabled = sdkDisabled

	return config
}

func getTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	otelConfig := parseOtelConfig()
	if otelConfig.endpoint == "" {
		logger.Info("OTEL_EXPORTER_OTLP_ENDPOINT not set, skipping OpenTelemetry tracing")
		return nil, nil
	}

	grpcOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(otelConfig.endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	}

	if otelConfig.insecure {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
	} else {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	return otlptracegrpc.New(ctx, grpcOpts...)
}

func InitTracer(serviceName, environment string, attributes map[string]string) (func(context.Context) error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	otelConfig := parseOtelConfig()

	if otelConfig.sdkDisabled {
		logger.Info("OpenTelemetry", "status", "disabled")
		tracer = trace.NewNoopTracerProvider().Tracer(serviceName)
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("deployment.environment", environment),
		),
	)
	if err != nil {
		return nil, err
	}

	for k, v := range attributes {
		res, _ = resource.Merge(res, resource.NewWithAttributes(semconv.SchemaURL, attribute.String(k, v)))
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	traceExporter, err := getTraceExporter(ctx)
	if err != nil {
		return nil, err
	}

	if traceExporter != nil {
		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		tracerProvider.RegisterSpanProcessor(bsp)
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	tracer = tracerProvider.Tracer(serviceName)

	logger.Info("OpenTelemetry", "status", "enabled")

	return func(ctx context.Context) error {
		if ctx.Err() != nil {
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			ctx = ctxWithTimeout
		}

		err := tracerProvider.Shutdown(ctx)
		if err != nil {
			logger.Error("Error shutting down trace provider", "error", err)
		}

		if traceExporter != nil {
			if err = traceExporter.Shutdown(ctx); err != nil {
				logger.Error("Error shutting down trace exporter", "error", err)
			}
		}

		return err
	}, nil
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}
