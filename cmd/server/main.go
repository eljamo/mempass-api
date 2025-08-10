package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/eljamo/mempass-api/internal/env"
	"github.com/eljamo/mempass-api/internal/interceptor"
	"github.com/eljamo/mempass-api/internal/tel"
)

func main() {
	ctx := context.Background()
	// logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logs, err := tel.InitLogging(ctx)
	if err != nil {
		panic(err)
	}
	defer logs.Shutdown(ctx)

	logger := logs.Logger

	err = run(ctx, logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type cfg struct {
	httpPort int
}

type application struct {
	config       cfg
	interceptors connect.Option
	logger       *slog.Logger
	wg           sync.WaitGroup
}

var defaultHTTPPort = 4321

func run(ctx context.Context, logger *slog.Logger) error {
	var cfg cfg
	cfg.httpPort = env.GetInt("HTTP_PORT", defaultHTTPPort)

	tracer, err := tel.InitTracing(ctx)
	if err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}
	defer tracer.Shutdown(ctx)

	meter, err := tel.InitMeter(ctx)
	if err != nil {
		logger.Error("failed to initialize metrics", "error", err)
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}
	defer meter.Shutdown(ctx)

	otel, err := otelconnect.NewInterceptor(
		otelconnect.WithTrustRemote(),
		otelconnect.WithoutServerPeerAttributes(),
	)
	if err != nil {
		logger.Error("failed to create OTEL interceptor", "error", err)
		return fmt.Errorf("failed to create OTEL interceptor: %w", err)
	}

	interceptors := connect.WithInterceptors(
		interceptor.NewRequestIDInterceptor(
			env.GetBool("ALLOW_EMPTY_REQUEST_ID", true),
			logger,
		),
		otel,
	)

	app := &application{
		config:       cfg,
		interceptors: interceptors,
		logger:       logger,
	}

	return app.serve()
}
