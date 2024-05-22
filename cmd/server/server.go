package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	defaultIdleTimeout          time.Duration = time.Minute
	defaultMaxHeaderBytes       int           = 8 * 1024
	defaultCompressionThreshold int           = 1024
	defaultReadHeaderTimeout    time.Duration = time.Second
	defaultReadTimeout          time.Duration = 5 * time.Second
	defaultShutdownPeriod       time.Duration = 30 * time.Second
	defaultWriteTimeout         time.Duration = 10 * time.Second
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", app.config.httpPort),
		Handler: h2c.NewHandler(
			app.routes(),
			&http2.Server{},
		),
		ErrorLog:          slog.NewLogLogger(app.logger.Handler(), slog.LevelWarn),
		IdleTimeout:       defaultIdleTimeout,
		MaxHeaderBytes:    defaultMaxHeaderBytes,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
	}

	shutdownErrorChan := make(chan error)

	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", slog.Group("server", "addr", srv.Addr))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	err = <-shutdownErrorChan
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	app.logger.Info("stopped server", slog.Group("server", "addr", srv.Addr))

	app.wg.Wait()
	return nil
}
