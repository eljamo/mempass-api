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

type errorStreamingClientInterceptor struct {
	connect.StreamingClientConn
	err error
}

type Interceptor struct {
	logger *slog.Logger
}

var _ connect.Interceptor = &Interceptor{}

func NewRequestIDInterceptor(logger *slog.Logger) *Interceptor {
	return &Interceptor{
		logger: logger,
	}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		err := i.validateHeader(request.Header())
		if err != nil {
			return nil, err
		}

		return next(ctx, request)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		err := i.validateHeader(conn.RequestHeader())
		if err != nil {
			return &errorStreamingClientInterceptor{
				StreamingClientConn: conn,
				err:                 connect.NewError(connect.CodeInternal, err),
			}
		}

		return conn
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		err := i.validateHeader(conn.RequestHeader())
		if err != nil {
			return err
		}

		return next(ctx, conn)
	}
}

func (i *Interceptor) validateHeader(headers http.Header) error {
	reqId := headers.Get(HeaderXRequestID)

	if reqId == "" {
		ulid, err := ulid.Generate()
		if err != nil {
			i.logger.Error("failed to set x-request-id", "err", err)

			return connect.NewError(
				connect.CodeInternal,
				errors.New("failed to set x-request-id"),
			)
		}
		headers.Set(HeaderXRequestID, ulid)

		i.logger.Info("set x-request-id header successfully", "id", ulid)
	} else {
		err := ulid.ValidateUlid(reqId)
		if err != nil {
			i.logger.Error("invalid x-request-id prodivded", "id", reqId)
			return connect.NewError(
				connect.CodeInternal,
				errors.New("invalid x-request-id provided"),
			)
		}

		i.logger.Info("validated x-request-id header successfully", "id", reqId)
	}

	return nil
}
