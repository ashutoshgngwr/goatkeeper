package goatkeeper

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/pkg/errors"
)

// MiddlewareOptions defines options for configuring the
// GoatKeeper middleware
type MiddlewareOptions struct {
	OpenAPISpec      []byte
	ValidateResponse bool // should goatkeeper also validate http responses
}

type middleware struct {
	*openapi3filter.Router
}

func (m *middleware) serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO validate request
		// TOOD write a dummy http.ResponseWriter
		next.ServeHTTP(w, r)
		// TODO validate data from dummy response writer
		// TODO write to actual writer
	})
}

// NewMiddleware creates a new HTTP middleware that will use OpenAPI Spec to
// validate all requests and responses.
func NewMiddleware(opts MiddlewareOptions) (func(http.Handler) http.Handler, error) {
	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(opts.OpenAPISpec)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse openapi spec")
	}

	m := middleware{
		Router: openapi3filter.NewRouter().WithSwagger(spec),
	}

	return m.serve, nil
}
