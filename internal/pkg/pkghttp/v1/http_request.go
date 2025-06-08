package pkghttp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type (
	Request interface {
		Decode(v any) error
		Header() http.Header
		URL() *url.URL
		Raw() *http.Request
	}

	RequestReadWriter interface {
		Decode(v any) error
		Encode(v any) error
		Header() http.Header
		URL() *url.URL
	}

	request struct {
		httpReq *http.Request
	}
)

func NewRequest(r *http.Request) *request { //nolint:revive // this is a factory function
	return &request{
		httpReq: r,
	}

}

func (r *request) Body() *reader {
	var err error
	var body *reader
	if r.httpReq.Body != nil {
		body, err = newReadWriter(r.httpReq.Body)
		if err != nil {
			return nil
		}

		r.httpReq.Body = body
	}

	return body
}

func (r *request) Decode(v any) error {
	if body := r.Body(); body != nil {
		b, err := io.ReadAll(r.Body())
		if err != nil {
			return err
		}

		return json.Unmarshal(b, v)
	}

	return nil
}

func (r *request) Encode(v any) error {
	if body := r.Body(); body != nil {
		return json.NewEncoder(body).Encode(v)
	}

	return nil
}

func (r *request) Header() http.Header {
	return r.httpReq.Header
}

func (r *request) URL() *url.URL {
	return r.httpReq.URL
}

func (r *request) Raw() *http.Request {
	return r.httpReq
}
