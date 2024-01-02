package main

import (
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

func run(logger *slog.Logger) error {
	var cfg cfg

	cfg.httpPort = env.GetInt("HTTP_PORT", 4321)

	interceptors := connect.WithInterceptors(
		interceptor.NewRequestIDInterceptor(logger),
		otelconnect.NewInterceptor(),
	)

	app := &application{
		config:       cfg,
		interceptors: interceptors,
		logger:       logger,
	}

	return app.serve()
}
