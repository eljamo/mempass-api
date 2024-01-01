package interceptor

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/eljamo/mempass-api/internal/ulid"
)

const requestIdHeader = "x-request-id"

func NewRequestIDInterceptor() connect.UnaryInterceptorFunc {

	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			if req.Header().Get(requestIdHeader) == "" {
				ulid, err := ulid.Generate()
				if err != nil {
					return nil, connect.NewError(
						connect.CodeInternal,
						errors.New("failed to generate x-request-id"),
					)
				}

				req.Header().Set(requestIdHeader, ulid)
			} else {
				err := ulid.ValidateUlid(req.Header().Get(requestIdHeader))
				if err != nil {
					return nil, connect.NewError(
						connect.CodeInternal,
						errors.New("invalid x-request-id provided"),
					)
				}
			}

			return next(ctx, req)
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
