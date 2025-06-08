package pkghttp

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	applicationJSON = "application/json; charset=utf-8"
	contentType     = "Content-Type"
)

// StatusCodeAware interface should be implemented by the response struct which desires status code.
type StatusCodeAware interface {
	StatusCode() int
}

// HeaderAware interface should be implemented by the response struct which desires additional headers.
type HeaderAware interface {
	Headers() http.Header
}

// ResponseEncoder function defines the contract for encoding successful responses.
type ResponseEncoder func(ctx context.Context, w http.ResponseWriter, response any) error

// ErrorResponseEncoder function defines the contract for encoding error responses.
type ErrorResponseEncoder func(ctx context.Context, err error, w http.ResponseWriter)

func DefaultResponseEncoder(_ context.Context, w http.ResponseWriter, response any) error {
	code := writeHeaderAndStatusCode(w, response)

	if code == http.StatusNoContent {
		return nil
	}

	return json.NewEncoder(w).Encode(response)
}

func writeHeaderAndStatusCode(w http.ResponseWriter, response any) int {
	w.Header().Set(contentType, applicationJSON)

	if headerAware, ok := response.(HeaderAware); ok {
		for k, values := range headerAware.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	code := http.StatusOK

	if sc, ok := response.(StatusCodeAware); ok {
		code = sc.StatusCode()
	}

	w.WriteHeader(code)

	return code
}

func DefaultErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(w).Encode(err.Error()) //nolint:errcheck,errchkjson // won't be an error
}
