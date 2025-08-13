package verify

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type VerifyConfig struct {
	CollectorEndpoint string        // e.g., "localhost:4317"
	Timeout           time.Duration // e.g., 5 * time.Second
}

func Run(cfg VerifyConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// --- TRACE ---
	traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("trace exporter init failed: %w", err)
	}
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp))
	otel.SetTracerProvider(tracerProvider)

	// --- METRIC ---
	metricExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(cfg.CollectorEndpoint), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("metric exporter init failed: %w", err)
	}
	reader := sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(2*time.Second))
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(meterProvider)

	// --- LOG ---
	logExp, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpoint(cfg.CollectorEndpoint), otlploggrpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("log exporter init failed: %w", err)
	}
	loggerProvider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)))
	global.SetLoggerProvider(loggerProvider)

	// Send test signals
	sendTrace(ctx, "verify-tracer")
	sendMetric(ctx, "verify-meter", "verify.counter")
	sendLog(ctx, "verify-logger")

	// Graceful shutdown
	_ = tracerProvider.Shutdown(ctx)
	_ = meterProvider.Shutdown(ctx)
	_ = loggerProvider.Shutdown(ctx)

	fmt.Println("✅ All signals sent")
	return nil
}

func sendTrace(ctx context.Context, name string) {
	tr := otel.Tracer(name)
	_, span := tr.Start(ctx, "test-span")
	span.SetAttributes(attribute.String("verify", "trace"))
	span.End()
	fmt.Println("Trace sent")
}

func sendMetric(ctx context.Context, meterName, counterName string) {
	meter := otel.Meter(meterName)
	counter, _ := meter.Int64Counter(counterName)
	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("verify", "metric")))
	fmt.Println("Metric sent")
}

func sendLog(ctx context.Context, loggerName string) {
	logger := global.GetLoggerProvider().Logger(loggerName)
	var record log.Record
	record.SetBody(log.StringValue("Verification log entry"))
	record.SetSeverity(log.SeverityInfo)
	record.AddAttributes(log.String("verify", "log"))
	logger.Emit(ctx, record)
	fmt.Println("Log sent")
}
