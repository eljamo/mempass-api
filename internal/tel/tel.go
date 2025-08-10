package tel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/eljamo/mempass-api/internal/env"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

var (
	metricsExporterEndpoint = env.GetString("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", env.OTLPExporterEndpoint)
	tracesExporterEndpoint  = env.GetString("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", env.OTLPExporterEndpoint)
	logsExporterEndpoint    = env.GetString("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", env.OTLPExporterEndpoint)
)

func newResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", env.ServiceName),
		),
	)
}

type Meter struct {
	exporter *otlpmetricgrpc.Exporter
	Provider *metricsdk.MeterProvider
}

func (m *Meter) exporterShutdown(ctx context.Context) error {
	return m.exporter.Shutdown(ctx)
}

func (m *Meter) Shutdown(ctx context.Context) error {
	err := m.exporterShutdown(ctx)
	if err != nil {
		return err
	}

	return m.Provider.Shutdown(ctx)
}

func InitMeter(ctx context.Context) (*Meter, error) {
	secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if env.OTLPExporterInsecure {
		secureOption = otlpmetricgrpc.WithInsecure()
	}

	exporter, err := otlpmetricgrpc.New(
		ctx,
		secureOption,
		otlpmetricgrpc.WithEndpoint(metricsExporterEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set resources: %w", err)
	}

	// Register the exporter with an SDK via a periodic reader.
	provider := metricsdk.NewMeterProvider(
		metricsdk.WithResource(res),
		metricsdk.WithReader(metricsdk.NewPeriodicReader(exporter)),
	)

	otel.SetMeterProvider(
		provider,
	)

	return &Meter{
		exporter: exporter,
		Provider: provider,
	}, nil
}

type Tracer struct {
	exporter *otlptrace.Exporter
	Provider *tracesdk.TracerProvider
}

func (t *Tracer) exporterShutdown(ctx context.Context) error {
	return t.exporter.Shutdown(ctx)
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	err := t.exporterShutdown(ctx)
	if err != nil {
		return err
	}

	return t.Provider.Shutdown(ctx)
}

func InitTracing(ctx context.Context) (*Tracer, error) {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if env.OTLPExporterInsecure {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(tracesExporterEndpoint),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set resources: %w", err)
	}

	provider := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(res),
	)

	otel.SetTracerProvider(
		provider,
	)

	return &Tracer{
		exporter: exporter,
		Provider: provider,
	}, nil
}

type Logger struct {
	exporter *otlploggrpc.Exporter
	Provider *logsdk.LoggerProvider
	Logger   *slog.Logger
}

func (l *Logger) exporterShutdown(ctx context.Context) error {
	return l.exporter.Shutdown(ctx)
}

func (l *Logger) Shutdown(ctx context.Context) error {
	err := l.exporterShutdown(ctx)
	if err != nil {
		return err
	}

	return l.Provider.Shutdown(ctx)
}

func InitLogging(ctx context.Context) (*Logger, error) {
	secureOption := otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if env.OTLPExporterInsecure {
		secureOption = otlploggrpc.WithInsecure()
	}

	exporter, err := otlploggrpc.New(
		ctx,
		secureOption,
		otlploggrpc.WithEndpoint(logsExporterEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to set resources: %w", err)
	}

	provider := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(log.NewBatchProcessor(exporter)),
		logsdk.WithResource(res),
	)

	global.SetLoggerProvider(provider)

	logger := otelslog.NewLogger(env.ServiceName, otelslog.WithLoggerProvider(provider))

	return &Logger{
		exporter: exporter,
		Provider: provider,
		Logger:   logger,
	}, nil
}
