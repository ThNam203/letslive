package tracer

import (
	"context"
	"errors"
	"sen1or/letslive/finance/config"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var MyTracer oteltrace.Tracer

func SetupOTelSDK(ctx context.Context, cfg config.Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	tracerProvider, err := newTracerProvider(ctx, cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	MyTracer = otel.Tracer(cfg.Service.Name)
	return
}

func newTracerProvider(ctx context.Context, cfg config.Config) (*trace.TracerProvider, error) {
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Tracer.Endpoint)}
	if !cfg.Tracer.Secure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Duration(cfg.Tracer.BatchTimeout)*time.Millisecond)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.Service.Name),
		)),
	), nil
}
