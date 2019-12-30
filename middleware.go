package goatkeeper

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-logr/logr"
	logrtesting "github.com/go-logr/logr/testing"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// MiddlewareResponse defines responses that middleware will
// write to http requests if it decides to not propagate it
// further.
type MiddlewareResponse struct {
	Status  int
	Body    []byte
	Headers http.Header
}

func (response *MiddlewareResponse) writeToHTTPResponseWriter(w http.ResponseWriter) {
	for key, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(response.Status)

	// TODO add error checking here
	w.Write(response.Body)
}

// MiddlewareOptions defines options for configuring the
// GoatKeeper middleware.
type MiddlewareOptions struct {
	Logger                  logr.Logger
	OpenAPISpec             []byte
	ValidateResponse        bool                // should goatkeeper also validate http responses
	InvalidRequestResponse  *MiddlewareResponse // response to write if request is invalid
	InvalidResponseResponse *MiddlewareResponse // response to write if response is invalid
}

// DefaultMiddlewareOptions defines default options used by the
// middleware.
var DefaultMiddlewareOptions = MiddlewareOptions{
	Logger:           logrtesting.NullLogger{},
	ValidateResponse: true,
	InvalidRequestResponse: &MiddlewareResponse{
		Status: http.StatusBadRequest,
	},
	InvalidResponseResponse: &MiddlewareResponse{
		Status: http.StatusInternalServerError,
	},
}

type middleware struct {
	*MiddlewareOptions
	*openapi3filter.Router
	context.Context
}

func (m *middleware) validateRequest(r *http.Request) error {
	route, pathParams, err := m.FindRoute(r.Method, r.URL)
	if err != nil {
		return err
	}

	return openapi3filter.ValidateRequest(m, &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	})
}

func (m *middleware) validateResponse(recorder *httpResponseRecorder) error {
	return openapi3filter.ValidateResponse(m, &openapi3filter.ResponseValidationInput{
		Status: recorder.status,
		Header: recorder.headers,
		Body:   ioutil.NopCloser(bytes.NewReader(recorder.body)),
	})
}

func (m *middleware) serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO create sub-logger with values.

		err := m.validateRequest(r)
		if err != nil {
			m.InvalidRequestResponse.writeToHTTPResponseWriter(w)
			m.Logger.Error(err, "invalid request data")
			return
		}

		if !m.ValidateResponse {
			next.ServeHTTP(w, r)
			return
		}

		recorder := &httpResponseRecorder{}
		next.ServeHTTP(recorder, r)

		err = m.validateResponse(recorder)
		if err != nil {
			m.InvalidResponseResponse.writeToHTTPResponseWriter(w)
			m.Logger.Error(err, "invalid response data")
			return
		}

		recorder.writeToHTTPResponseWriter(w)
	})
}

// NewMiddleware creates a new HTTP middleware that will use OpenAPI Spec to
// validate all requests and responses.
func NewMiddleware(opts *MiddlewareOptions) (func(http.Handler) http.Handler, error) {
	err := mergo.Merge(opts, DefaultMiddlewareOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable merge given options with defaults")
	}

	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(opts.OpenAPISpec)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse openapi spec")
	}

	m := middleware{
		Context:           context.TODO(),
		MiddlewareOptions: opts,
		Router:            openapi3filter.NewRouter().WithSwagger(spec),
	}

	return m.serve, nil
}
