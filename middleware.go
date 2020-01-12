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

// MiddlewareResponse defines a response format that middleware will
// write to http requests if their request body or response do not
// adhere to defined OpenAPI specification.
type MiddlewareResponse struct {
	Status  int
	Body    []byte
	Headers http.Header
}

// writeToResponse writes the response to an http.ResponseWriter.
func (response *MiddlewareResponse) writeToResponse(w http.ResponseWriter) error {
	for key, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(response.Status)
	_, err := w.Write(response.Body)
	return err
}

// MiddlewareOptions defines options for configuring the GoatKeeper middleware.
type MiddlewareOptions struct {
	Logger                  logr.Logger
	OpenAPISpec             []byte
	ValidateResponse        bool                // should goatkeeper also validate http responses
	InvalidRequestResponse  *MiddlewareResponse // response to write if request is invalid
	InvalidResponseResponse *MiddlewareResponse // response to write if response is invalid
}

// DefaultMiddlewareOptions defines default options used by the middleware.
// These options are only used if a option isn't specified when initializing
// the middleware.
var DefaultMiddlewareOptions = MiddlewareOptions{
	Logger:           logrtesting.NullLogger{},
	ValidateResponse: false,
	InvalidRequestResponse: &MiddlewareResponse{
		Headers: http.Header{},
		Status:  http.StatusBadRequest,
		Body:    []byte{},
	},
	InvalidResponseResponse: &MiddlewareResponse{
		Status:  http.StatusInternalServerError,
		Headers: http.Header{},
		Body:    []byte{},
	},
}

// middleware implements the goatkeeper middleware.
type middleware struct {
	*MiddlewareOptions
	*openapi3filter.Router
	context.Context
}

// validateRequest validates the given request with data from the parsed
// OpenAPI specification.
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

// validateResponse validates the given response with data from the parsed
// OpenAPI specification.
func (m *middleware) validateResponse(recorder *httpResponseRecorder, r *http.Request) error {
	route, pathParams, err := m.FindRoute(r.Method, r.URL)
	if err != nil {
		return err
	}

	return openapi3filter.ValidateResponse(m, &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
		},
		Status: recorder.Status,
		Header: recorder.Headers,
		Body:   ioutil.NopCloser(bytes.NewBuffer(recorder.Body.Bytes())),
	})
}

// serve is the actual goatkeeper middleware.
func (m *middleware) serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := m.Logger.WithValues("Method", r.Method, "URL", r.URL)

		err := m.validateRequest(r)
		if err != nil {
			logger.Error(err, "invalid request data")
			if err = m.InvalidRequestResponse.writeToResponse(w); err != nil {
				logger.Error(err, "unable to write response")
			}

			return
		}

		if !m.ValidateResponse {
			next.ServeHTTP(w, r)
			return
		}

		recorder := newHTTPResponseRecorder()
		next.ServeHTTP(recorder, r)

		err = m.validateResponse(recorder, r)
		if err != nil {
			logger.Error(err, "invalid response data")
			if err = m.InvalidResponseResponse.writeToResponse(w); err != nil {
				logger.Error(err, "unable to write response")
			}
			return
		}

		if err = recorder.writeToResponse(w); err != nil {
			logger.Error(err, "unable to write response")
		}
	})
}

// NewMiddleware creates a new HTTP middleware that will use the given
// OpenAPI Spec to validate all requests and their responses.
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
