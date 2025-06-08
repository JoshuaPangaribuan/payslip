package pkghttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type responseEncoderTestArgs struct {
	endpoint *Endpoint
}

type responseEncoderTest struct {
	name                string
	args                responseEncoderTestArgs
	expectedCode        int
	expectedBody        string
	expectedHeadersFunc func(header http.Header)
}

func Test_DefaultResponseEncoder(t *testing.T) {
	tests := []responseEncoderTest{
		{
			name: "success without body",
			args: responseEncoderTestArgs{
				endpoint: &Endpoint{
					handler: func(ctx context.Context, r Request) (any, error) {
						return nil, nil
					},
					responseEncoder:      DefaultResponseEncoder,
					errorResponseEncoder: DefaultErrorEncoder,
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "null\n",
			expectedHeadersFunc: func(header http.Header) {
				header.Add(contentType, applicationJSON)
			},
		},
		{
			name: "success with body",
			args: responseEncoderTestArgs{
				endpoint: &Endpoint{
					handler: func(ctx context.Context, r Request) (any, error) {
						return "body", nil
					},
					responseEncoder:      DefaultResponseEncoder,
					errorResponseEncoder: DefaultErrorEncoder,
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "\"body\"\n",
			expectedHeadersFunc: func(header http.Header) {
				header.Add(contentType, applicationJSON)
			},
		},
		{
			name: "success with custom code & additional headers",
			args: responseEncoderTestArgs{
				endpoint: &Endpoint{
					handler: func(ctx context.Context, r Request) (any, error) {
						return customCodeAndHeaderDummy{}, nil
					},
					responseEncoder:      DefaultResponseEncoder,
					errorResponseEncoder: DefaultErrorEncoder,
				},
			},
			expectedCode: http.StatusAccepted,
			expectedBody: "{}\n",
			expectedHeadersFunc: func(header http.Header) {
				header.Add(contentType, applicationJSON)
				header.Add("additional", "dummy")
			},
		},
		{
			name: "success with no content",
			args: responseEncoderTestArgs{
				endpoint: &Endpoint{
					handler: func(ctx context.Context, r Request) (any, error) {
						return noResponseDummy{}, nil
					},
					responseEncoder:      DefaultResponseEncoder,
					errorResponseEncoder: DefaultErrorEncoder,
				},
			},
			expectedCode: http.StatusNoContent,
			expectedBody: "",
			expectedHeadersFunc: func(header http.Header) {
				header.Add(contentType, applicationJSON)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertResponseEncoderTest(t, test)
		})
	}
}

func Test_ErrorResponseEncoder(t *testing.T) {
	tests := []responseEncoderTest{
		{
			name: "error response",
			args: responseEncoderTestArgs{
				endpoint: &Endpoint{
					handler: func(ctx context.Context, r Request) (any, error) {
						return nil, errors.New("some error")
					},
					responseEncoder:      DefaultResponseEncoder,
					errorResponseEncoder: DefaultErrorEncoder,
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "\"some error\"\n",
			expectedHeadersFunc: func(header http.Header) {
				header.Add(contentType, applicationJSON)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertResponseEncoderTest(t, test)
		})
	}
}

func assertResponseEncoderTest(t *testing.T, test responseEncoderTest) {
	assertion := assert.New(t)
	request, _ := http.NewRequest("POST", "/example", nil)
	response := httptest.NewRecorder()
	expectedHeader := httptest.NewRecorder().Header()
	test.expectedHeadersFunc(expectedHeader)

	test.args.endpoint.ServeHTTP(response, request)

	assertion.Equal(test.expectedCode, response.Code)
	assertion.Equal(test.expectedBody, response.Body.String())
	assertion.Equal(expectedHeader, response.Header())
}

type noResponseDummy struct{}

func (noResponseDummy) StatusCode() int {
	return http.StatusNoContent
}

type customCodeAndHeaderDummy struct{}

func (customCodeAndHeaderDummy) Headers() http.Header {
	return map[string][]string{"additional": {"dummy"}}
}

func (customCodeAndHeaderDummy) StatusCode() int {
	return http.StatusAccepted
}
