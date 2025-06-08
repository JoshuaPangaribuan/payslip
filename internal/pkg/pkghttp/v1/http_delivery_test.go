package pkghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Server_Serve(t *testing.T) {
	type exampleBody struct {
		Example string `json:"example"`
	}

	type fields struct {
		responseEncoder       ResponseEncoder
		errorResponseEncoder  ErrorResponseEncoder
		requestDecoders       []RequestDecoder
		preRequestMiddlewares []PreRequestMiddleware
	}
	type args struct {
		body any
		e    EndpointHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "success",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response any) error {
					return nil
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
			},
			args: args{
				e: func(ctx context.Context, r Request) (any, error) {
					return nil, nil
				},
			},
		},
		{
			name: "success with request decoder",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response any) error {
					return nil
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
				requestDecoders: []RequestDecoder{WithPopulateContextFromHeader},
			},
			args: args{
				e: func(ctx context.Context, r Request) (any, error) {
					return nil, nil
				},
			},
		},
		{
			name: "endpoint error",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response any) error {
					return nil
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
				requestDecoders: []RequestDecoder{},
				preRequestMiddlewares: []PreRequestMiddleware{
					func(next EndpointHandler) EndpointHandler {
						return next
					},
				},
			},
			args: args{
				body: exampleBody{
					Example: "example",
				},
				e: func(ctx context.Context, r Request) (any, error) {
					type snapRequest struct {
						PartnerID string `json:"partner_id"`
					}

					var req snapRequest

					if err := r.Decode(&req); err != nil {
						return nil, err

					}

					fmt.Printf("req: %v\n", req)
					return nil, errors.New("test endpoint error")
				},
			},
		},
		{
			name: "encode response error",
			fields: fields{
				responseEncoder: func(ctx context.Context, w http.ResponseWriter, response any) error {
					return errors.New("error while encode response")
				},
				errorResponseEncoder: func(ctx context.Context, err error, w http.ResponseWriter) {
				},
			},
			args: args{
				e: func(ctx context.Context, r Request) (any, error) {
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(
				WithResponseEncoder(tt.fields.responseEncoder),
				WithErrorResponseEncoder(tt.fields.errorResponseEncoder),
				WithPreRequestMiddlewares(tt.fields.preRequestMiddlewares...),
			)

			opts := []EndpointOption{}

			for _, dec := range tt.fields.requestDecoders {
				opts = append(opts, WithRequestDecoder(dec))
			}

			e := s.Serve(tt.args.e, opts...)

			var req *http.Request
			var err error
			if tt.args.body != nil {
				bodyByte, err := json.Marshal(tt.args.body)
				assert.NoError(t, err)
				bodyBuffer := bytes.NewBuffer(bodyByte)

				req, err = http.NewRequest("POST", "/example", bodyBuffer)
			} else {
				req, err = http.NewRequest("POST", "/example", nil)

			}

			if err != nil {
				assert.Error(t, err)
			}

			rr := httptest.NewRecorder()
			e.ServeHTTP(rr, req)
		})
	}
}
