package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/eljamo/mempass-api/internal/env"
	"github.com/eljamo/mempass-api/internal/interceptor"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
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

func run(logger *slog.Logger) error {
	var cfg cfg
	cfg.httpPort = env.GetInt("HTTP_PORT", defaultHTTPPort)

	otel, err := otelconnect.NewInterceptor()
	if err != nil {
		return fmt.Errorf("failed to create OpenTelemetry interceptor: %w", err)
	}

	interceptors := connect.WithInterceptors(
		interceptor.NewRequestIDInterceptor(
			env.GetBool("ALLOW_EMPTY_REQUEST_ID", false),
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
