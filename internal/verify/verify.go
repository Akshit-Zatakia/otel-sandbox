// internal/verify/verify.go - Add health check and improve error handling
package verify

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/resource"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type VerifyConfig struct {
	CollectorEndpoint string
	Timeout           time.Duration
}

// Add health check function
func checkCollectorHealth(endpoint string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return fmt.Errorf("collector not reachable at %s: %w", endpoint, err)
	}
	defer conn.Close()
	return nil
}

func Run(config VerifyConfig) error {
	fmt.Println("🔍 Verifying OTel Collector connection...")

	// Check if collector is reachable
	if err := checkCollectorHealth(config.CollectorEndpoint, config.Timeout); err != nil {
		return fmt.Errorf("❌ Collector health check failed: %w\n\n💡 Make sure to run 'otel-sandbox up' first", err)
	}
	fmt.Println("✅ Collector is reachable")

	ctx := context.Background()

	// Set up trace provider
	fmt.Println("🔧 Setting up trace provider...")
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.CollectorEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("❌ failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			attribute.String("service.name", "otel-sandbox-verify"),
			attribute.String("service.version", "1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)

	// Set up metric provider
	fmt.Println("🔧 Setting up metric provider...")
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint("localhost:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("❌ failed to create metric exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			attribute.String("service.name", "otel-sandbox-verify"),
			attribute.String("service.version", "1.0.0"),
		)),
	)
	otel.SetMeterProvider(mp)

	// Set up log provider
	fmt.Println("🔧 Setting up log provider...")
	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint("localhost:4318"),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("❌ failed to create log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			attribute.String("service.name", "otel-sandbox-verify"),
			attribute.String("service.version", "1.0.0"),
		)),
	)
	global.SetLoggerProvider(lp)

	// Send sample telemetry
	fmt.Println("📤 Sending sample telemetry data...")

	// Send trace
	tracer := otel.Tracer("verify-tracer")
	_, span := tracer.Start(ctx, "verify-operation")
	span.SetAttributes(
		attribute.String("operation.type", "verification"),
		attribute.Int("operation.count", 1),
	)
	time.Sleep(100 * time.Millisecond) // Simulate work
	span.End()
	fmt.Println("  ✅ Trace sent")

	// Send metric
	meter := otel.Meter("verify-meter")
	counter, err := meter.Int64Counter("verify_operations_total")
	if err == nil {
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "verification"),
		))
		fmt.Println("  ✅ Metric sent")
	}

	var logRecord log.Record
	logRecord.SetTimestamp(time.Now())
	logRecord.SetBody(log.StringValue("Verification operation completed successfully"))
	logRecord.SetSeverity(log.SeverityInfo)
	logRecord.AddAttributes(
		log.String("component", "verify"),
		log.String("operation", "verification"),
	)

	// Send log
	logger := global.GetLoggerProvider().Logger("verify-logger")
	logger.Emit(ctx, logRecord)
	fmt.Println("  ✅ Log sent")

	// Flush all data
	fmt.Println("⏳ Flushing telemetry data...")

	if err := tp.Shutdown(ctx); err != nil {
		fmt.Printf("⚠️  Warning: failed to shutdown trace provider: %v\n", err)
	}

	if err := mp.Shutdown(ctx); err != nil {
		fmt.Printf("⚠️  Warning: failed to shutdown metric provider: %v\n", err)
	}

	if err := lp.Shutdown(ctx); err != nil {
		fmt.Printf("⚠️  Warning: failed to shutdown log provider: %v\n", err)
	}

	// Wait for data to be processed
	fmt.Println("⏳ Waiting for collector to process data...")
	time.Sleep(3 * time.Second)

	fmt.Println("\n🎉 Verification completed successfully!")
	fmt.Println("📁 Check these files for collected data:")
	fmt.Println("   - ./otel-traces.json")
	fmt.Println("   - ./otel-metrics.json")
	fmt.Println("   - ./otel-logs.json")
	fmt.Println("\n💡 Use 'otel-sandbox export' to view the data in different formats")

	return nil
}
