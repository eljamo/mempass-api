package main

import (
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/eljamo/mempass-api/internal/env"
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
	config cfg
	logger *slog.Logger
	wg     sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg cfg

	cfg.httpPort = env.GetInt("HTTP_PORT", 4321)

	app := &application{
		config: cfg,
		logger: logger,
	}

	return app.serve()
}
