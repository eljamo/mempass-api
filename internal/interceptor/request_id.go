package interceptor

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	connect "connectrpc.com/connect"
	"github.com/eljamo/mempass-api/internal/ulid"
)

const HeaderXRequestID = "x-request-id"

var (
	ErrorRequestIDMissing     = errors.New("x-request-id header is missing")
	ErrorFailedToSetRequestID = errors.New("failed to set x-request-id header")
	ErrorInvalidRequestID     = errors.New("invalid x-request-id provided")
)

type errorStreamingClientInterceptor struct {
	connect.StreamingClientConn
	err error
}

type Interceptor struct {
	allowEmptyRequestID bool
	logger              *slog.Logger
}

var _ connect.Interceptor = &Interceptor{}

// NewRequestIDInterceptor creates a new Interceptor instance.
func NewRequestIDInterceptor(allowEmptyRequestID bool, logger *slog.Logger) *Interceptor {
	return &Interceptor{
		allowEmptyRequestID: allowEmptyRequestID,
		logger:              logger,
	}
}

// WrapUnary wraps a unary function with request ID validation.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		if err := i.validateHeader(request.Header()); err != nil {
			return nil, err
		}
		return next(ctx, request)
	}
}

// WrapStreamingClient wraps a streaming client function with request ID validation.
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		if err := i.validateHeader(conn.RequestHeader()); err != nil {
			return &errorStreamingClientInterceptor{
				StreamingClientConn: conn,
				err:                 connect.NewError(connect.CodeInternal, err),
			}
		}
		return conn
	}
}

// WrapStreamingHandler wraps a streaming handler function with request ID validation.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if err := i.validateHeader(conn.RequestHeader()); err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

// handleEmptyHeader handles cases where the x-request-id header is missing.
func (i *Interceptor) handleEmptyHeader(headers http.Header) error {
	if i.allowEmptyRequestID {
		ulid, err := ulid.Generate()
		if err != nil {
			i.logger.Error("failed to set x-request-id header", slog.String("error", ErrorFailedToSetRequestID.Error()))
			return connect.NewError(connect.CodeInternal, ErrorFailedToSetRequestID)
		}
		headers.Set(HeaderXRequestID, ulid)
		i.logger.Info("set x-request-id header successfully", slog.String("id", ulid))
	} else {
		i.logger.Error("x-request-id header is missing")
		return connect.NewError(connect.CodeInternal, ErrorRequestIDMissing)
	}

	return nil
}

// validateHeader validates and sets the x-request-id header.
func (i *Interceptor) validateHeader(headers http.Header) error {
	reqId := headers.Get(HeaderXRequestID)

	if reqId == "" {
		return i.handleEmptyHeader(headers)
	} else {
		if err := ulid.ValidateUlid(reqId); err != nil {
			i.logger.Error("invalid x-request-id provided", slog.String("id", reqId))
			return connect.NewError(connect.CodeInternal, ErrorInvalidRequestID)
		}
		i.logger.Info("validated x-request-id header successfully", slog.String("id", reqId))
	}

	return nil
}
