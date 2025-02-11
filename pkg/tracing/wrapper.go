package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

type SpanWrapper struct {
	trace.Span
}

func NewSpanWrapper(ctx context.Context, name string) (context.Context, *SpanWrapper) {
	// Start a new span
	ctx, span := otel.Tracer("passage-server").Start(ctx, name)

	// Wrap the span in a CustomSpan and add standard attributes
	customSpan := &SpanWrapper{Span: span}
	customSpan.WithStandardAttributes(ctx)

	return ctx, customSpan
}

// WithStandardAttributes adds standard attributes like user ID, request ID, etc.
func (s *SpanWrapper) WithStandardAttributes(ctx context.Context) *SpanWrapper {
	// Retrieve baggage from the context
	bag := baggage.FromContext(ctx)

	// Iterate over all baggage members and add them as span attributes
	for _, member := range bag.Members() {
		s.SetAttributes(attribute.String(member.Key(), member.Value()))
	}
	return s
}

func (s *SpanWrapper) LogError(err error) {
	if err != nil {
		s.RecordError(err)
		s.SetAttributes(attribute.String("error.message", err.Error()))
	}
}

func (s *SpanWrapper) GetTraceID() string {
	spanContext := s.SpanContext()

	// Check if the span context contains a valid trace ID
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}

	return ""
}
