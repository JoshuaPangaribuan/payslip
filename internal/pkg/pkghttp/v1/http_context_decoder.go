package pkghttp

import (
	"context"
)

type contextKey int

const (
	// ContextKeyXPartnerID represents the context key for the XPartnerID value.
	ContextKeyAuthorization contextKey = iota
)

func WithPopulateContextFromHeader(
	ctx context.Context,
	r RequestReadWriter,
) (context.Context, error) {
	for k, v := range map[contextKey]string{
		ContextKeyAuthorization: r.Header().Get("Authorization"),
	} {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx, nil
}
