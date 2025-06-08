package pkghttp

import (
	"context"
	"net/http"
)

type (
	EndpointHandler func(ctx context.Context, r Request) (response interface{}, err error)

	PreRequestMiddleware func(next EndpointHandler) EndpointHandler

	Endpoint struct {
		handler              EndpointHandler
		responseEncoder      ResponseEncoder
		requestDecoders      []RequestDecoder
		errorResponseEncoder ErrorResponseEncoder
		middlewares          []PreRequestMiddleware
	}

	EndpointOption func(*Endpoint)
)

func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := &request{
		httpReq: r,
	}

	if e.requestDecoders != nil {
		for _, rd := range e.requestDecoders {
			var err error
			ctx, err = rd(ctx, req)
			if err != nil {
				e.errorResponseEncoder(ctx, err, w)

				return
			}
		}
	}

	handler := e.handler
	for _, m := range e.middlewares {
		handler = m(handler)
	}

	res, err := handler(ctx, req)
	if err != nil {
		e.errorResponseEncoder(ctx, err, w)

		return
	}

	if err := e.responseEncoder(ctx, w, res); err != nil {
		e.errorResponseEncoder(ctx, err, w)

		return
	}
}

func WithRequestDecoder(decoder RequestDecoder) EndpointOption {
	return func(endpoint *Endpoint) {
		endpoint.requestDecoders = append(endpoint.requestDecoders, decoder)
	}
}

func WithEndpointResponseEncoder(encoder ResponseEncoder) EndpointOption {
	return func(endpoint *Endpoint) {
		endpoint.responseEncoder = encoder
	}
}

func WithEndpointErrorResponseEncoder(encoder ErrorResponseEncoder) EndpointOption {
	return func(endpoint *Endpoint) {
		endpoint.errorResponseEncoder = encoder
	}
}

func WithPreRequestMiddleware(middlewares ...PreRequestMiddleware) EndpointOption {
	return func(endpoint *Endpoint) {
		endpoint.middlewares = append(endpoint.middlewares, middlewares...)
	}
}
