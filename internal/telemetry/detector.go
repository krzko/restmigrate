package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func DetectEnvironment() (*resource.Resource, error) {
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("restmigrate"),
			attribute.String("deployment.environment", getEnvironment()),
		),
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "unknown"
	}
	return env
}
