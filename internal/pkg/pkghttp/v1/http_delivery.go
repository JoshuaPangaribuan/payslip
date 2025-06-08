package pkghttp

type Server struct {
	responseEncoder      ResponseEncoder
	errorResponseEncoder ErrorResponseEncoder
	middlewares          []PreRequestMiddleware
}

func NewServer(options ...ServerOption) *Server {
	s := &Server{}

	defaultServer(s)

	for _, o := range options {
		o.Apply(s)
	}

	return s
}

func (s *Server) Serve(handler EndpointHandler, options ...EndpointOption) *Endpoint {
	endpoint := &Endpoint{
		handler:              handler,
		responseEncoder:      s.responseEncoder,
		errorResponseEncoder: s.errorResponseEncoder,
		middlewares:          s.middlewares,
	}

	for _, option := range options {
		option(endpoint)
	}

	return endpoint
}

func defaultServer(s *Server) {
	s.responseEncoder = DefaultResponseEncoder
	s.errorResponseEncoder = DefaultErrorEncoder
}

type ServerOption interface {
	Apply(*Server)
}

type ServerOptionFunc func(*Server)

func (o ServerOptionFunc) Apply(s *Server) {
	o(s)
}

func WithResponseEncoder(e ResponseEncoder) ServerOption {
	return ServerOptionFunc(func(s *Server) {
		s.responseEncoder = e
	})
}

func WithErrorResponseEncoder(e ErrorResponseEncoder) ServerOption {
	return ServerOptionFunc(func(s *Server) {
		s.errorResponseEncoder = e
	})
}

func WithPreRequestMiddlewares(middlewares ...PreRequestMiddleware) ServerOption {
	return ServerOptionFunc(func(s *Server) {
		s.middlewares = append(s.middlewares, middlewares...)
	})
}
