package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const tracerName = "github.com/fleetml/fleetml/server"

// Config holds tracing configuration.
type Config struct {
	Enabled     bool
	Endpoint    string  // OTLP HTTP endpoint (e.g., "localhost:4318")
	ServiceName string  // defaults to "fleetml-server"
	SampleRate  float64 // 0.0–1.0, defaults to 1.0
}

// Provider wraps the OpenTelemetry TracerProvider and manages lifecycle.
type Provider struct {
	tp     *sdktrace.TracerProvider
	logger *zap.SugaredLogger
}

// Init sets up OpenTelemetry tracing. Returns a Provider whose Shutdown()
// must be called on application exit. If cfg.Enabled is false, returns a
// no-op provider that still exposes the same API surface.
func Init(ctx context.Context, cfg Config, version string, logger *zap.SugaredLogger) (*Provider, error) {
	if !cfg.Enabled {
		logger.Info("tracing disabled")
		return &Provider{logger: logger}, nil
	}

	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "fleetml-server"
	}

	sampleRate := cfg.SampleRate
	if sampleRate <= 0 {
		sampleRate = 1.0
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Infow("tracing initialized", "endpoint", cfg.Endpoint, "sample_rate", sampleRate)
	return &Provider{tp: tp, logger: logger}, nil
}

// Shutdown flushes pending spans and shuts down the tracer provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tp == nil {
		return nil
	}
	p.logger.Info("shutting down tracer provider")
	return p.tp.Shutdown(ctx)
}

// Tracer returns a named tracer for creating spans.
func Tracer() trace.Tracer {
	return otel.Tracer(tracerName)
}

// HTTPMiddleware creates an HTTP middleware that starts a span for each request,
// records HTTP attributes, and propagates the trace context.
func HTTPMiddleware(next http.Handler) http.Handler {
	tracer := Tracer()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract incoming trace context
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		spanName := r.Method + " " + r.URL.Path
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPTargetKey.String(r.URL.Path),
				attribute.String("http.client_ip", r.RemoteAddr),
			),
		)
		defer span.End()

		// Wrap response writer to capture status code
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, r.WithContext(ctx))

		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(sw.statusCode))
		if sw.statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	})
}

// SpanFromContext creates a child span from the given context.
// Use this inside handlers and service methods to trace sub-operations.
func SpanFromContext(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.written {
		sw.statusCode = code
		sw.written = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.written {
		sw.written = true
	}
	return sw.ResponseWriter.Write(b)
}
