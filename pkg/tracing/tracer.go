package tracing

import (
	"context"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Config = config.GetConfig()
var tracerProvider *sdktrace.TracerProvider

// NewTracer initializes the tracer directly
func NewTracer() (*sdktrace.TracerProvider, error) {

	// Create a context for the tracing setup
	ctx := context.Background()

	// Set up the resource (e.g., service name, environment)
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(Config.Tracing.ServiceName),
			semconv.DeploymentEnvironment(Config.Tracing.EnvironmentName),
		),
	)

	var traceExporter sdktrace.SpanExporter
	var err error

	switch Config.Tracing.ConnectionType {
	case "http":
		traceExporter, err = otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(Config.Tracing.URL))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OTLP HTTP trace exporter: %w", err)
		}
	case "grpc":
		// Create a gRPC connection to the telemetry collector
		conn, err := grpc.NewClient(Config.Tracing.URL,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		// Initialize the OTLP gRPC trace exporter
		traceExporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OTLP gRPC trace exporter: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported connection type. Supported values are: grpc, http")
	}

	// Create and set up the tracer provider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(r),
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Set the text map propagator for baggage propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider, nil
}

// GetTracer returns the globally initialized tracer provider
func GetTracer() *sdktrace.TracerProvider {
	if tracerProvider == nil {
		// If tracer hasn't been initialized, return an error
		fmt.Println("Tracer provider not initialized")
		return nil
	}
	return tracerProvider
}

func NewTracingMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(Config.Tracing.ServiceName)
}

func AddBaggage(ctx context.Context, key string, value string) context.Context {
	// Create a new baggage member
	member, _ := baggage.NewMember(key, value)
	bag, _ := baggage.New(member)

	// Add the baggage to the context
	return baggage.ContextWithBaggage(ctx, bag)
}

func GetTraceIDFromSpan(span trace.Span) string {
	spanContext := span.SpanContext()

	// Check if the span context contains a valid trace ID
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}

	return ""
}
