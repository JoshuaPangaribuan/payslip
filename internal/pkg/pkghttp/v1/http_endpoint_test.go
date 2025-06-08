package pkghttp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EndpointOption(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  *Endpoint
		option    EndpointOption
		assertion func(*Endpoint) bool
	}{
		{
			name: "default",
			endpoint: &Endpoint{
				handler: func(ctx context.Context, r Request) (response interface{}, err error) {
					return nil, nil
				},
				responseEncoder:      DefaultResponseEncoder,
				errorResponseEncoder: DefaultErrorEncoder,
				middlewares:          nil,
			},
			option: func(endpoint *Endpoint) {

			},
			assertion: func(endpoint *Endpoint) bool {
				return endpoint.responseEncoder != nil &&
					endpoint.errorResponseEncoder != nil &&
					endpoint.handler != nil
			},
		},
		{
			name: "with endpoint response encoder",
			endpoint: &Endpoint{
				handler: func(ctx context.Context, r Request) (response interface{}, err error) {
					return nil, nil
				},
				responseEncoder:      DefaultResponseEncoder,
				errorResponseEncoder: DefaultErrorEncoder,
				middlewares:          nil,
			},
			option: WithEndpointResponseEncoder(nil),
			assertion: func(endpoint *Endpoint) bool {
				return endpoint.responseEncoder == nil
			},
		},
		{
			name: "with endpoint error response encoder",
			endpoint: &Endpoint{
				handler: func(ctx context.Context, r Request) (response interface{}, err error) {
					return nil, nil
				},
				responseEncoder:      DefaultResponseEncoder,
				errorResponseEncoder: DefaultErrorEncoder,
				middlewares:          nil,
			},
			option: WithEndpointErrorResponseEncoder(nil),
			assertion: func(endpoint *Endpoint) bool {
				return endpoint.errorResponseEncoder == nil
			},
		},
		{
			name: "with endpoint middlewares",
			endpoint: &Endpoint{
				handler: func(ctx context.Context, r Request) (response interface{}, err error) {
					return nil, nil
				},
				responseEncoder:      DefaultResponseEncoder,
				errorResponseEncoder: DefaultErrorEncoder,
				requestDecoders: []RequestDecoder{
					WithPopulateContextFromHeader,
				},
				middlewares: nil,
			},
			option: WithPreRequestMiddleware(
				func(next EndpointHandler) EndpointHandler {
					return func(ctx context.Context, r Request) (response interface{}, err error) {
						return next(ctx, r)
					}
				},
			),
			assertion: func(endpoint *Endpoint) bool {
				return len(endpoint.middlewares) == 1
			},
		},
		{
			name: "with request decoder",
			endpoint: &Endpoint{
				handler: func(ctx context.Context, r Request) (response interface{}, err error) {
					return nil, nil
				},
				responseEncoder:      DefaultResponseEncoder,
				errorResponseEncoder: DefaultErrorEncoder,
				requestDecoders: []RequestDecoder{
					WithPopulateContextFromHeader,
				},
				middlewares: nil,
			},
			option: WithRequestDecoder(WithPopulateContextFromHeader),
			assertion: func(endpoint *Endpoint) bool {
				return endpoint.requestDecoders != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.option(tt.endpoint)

			assert.True(t, tt.assertion(tt.endpoint))
		})
	}
}
