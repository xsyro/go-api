package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xsyro/goapi/config"
	"github.com/xsyro/goapi/internal/app/api/handler"
	"github.com/xsyro/goapi/internal/app/repo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Initializes an OTLP exporter, and configures the corresponding trace
func installExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	// Set up a trace exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithInsecure())
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	newResource := func() *resource.Resource {
		return resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-api"),
			semconv.ServiceVersion("0.0.1"),
		)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	// bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource()),
		// sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// setup W3C trace context as global propagator
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Setup Tracing
	ctx := context.Background()
	// Registers a tracer Provider globally.
	shutdown, err := installExportPipeline(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	// Setup Db connection
	cfg := config.AppConfig()
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Db.Host, cfg.Db.Port, cfg.Db.User, cfg.Db.DbName, cfg.Db.Pass, cfg.Db.SSLMode)
	repository, err := repo.NewRepo(dsn)
	if err != nil {
		log.Fatalf("failed connecting to postgres database: %v", err)
	}

	// Setup Redis
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{Addr: addr})

	h := handler.NewHandler(*logger, repository, redisClient, cfg)
	srv := http.Server{
		Addr:    ":8080",
		Handler: h,

		// These values are here to make sure that the server doesn't hang
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		// This value is extremely important, it prevents us from suffering a Slowloris attack
		IdleTimeout: time.Second * 60,
	}

	// Create a channel that listens on incomming interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	// Graceful shutdown
	go func() {
		// Wait for a new signal on channel
		<-signalChan
		// Signal received, shutdown the server
		fmt.Println("shutting down..")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		srv.Shutdown(ctx)

		// Check if context timeouts, in worst case call cancel via defer
		select {
		case <-time.After(21 * time.Second):
			fmt.Println("not all connections done")
		case <-ctx.Done():
		}
	}()

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server crashed: %v", err)
	}
}
