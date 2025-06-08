package pkghttp

import (
	"context"
	"net/http"
)

// RequestDecoder function defines the contract for decoding request.
type RequestDecoder func(ctx context.Context, r RequestReadWriter) (context.Context, error)

func DefaultRequestDecoder(ctx context.Context, _ *http.Request) (context.Context, error) {
	return ctx, nil
}
