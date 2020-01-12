package goatkeeper

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareResponse_writeToResponse(t *testing.T) {
	response := MiddlewareResponse{
		Status: 200,
		Body:   []byte("test"),
		Headers: http.Header{
			"Test": []string{"test"},
		},
	}
	recorder := httptest.NewRecorder()
	err := response.writeToResponse(recorder)

	assert.NoError(t, err)
	assert.Equal(t, response.Headers, recorder.Result().Header)
	assert.Equal(t, response.Status, recorder.Result().StatusCode)

	body, err := ioutil.ReadAll(recorder.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, response.Body, body)
}

func TestMiddleware_validateRequest(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      responses:
        '200':
          description: "test"
`,
		))

		assert.NoError(t, err)
		m := middleware{
			Context: context.TODO(),
			Router:  openapi3filter.NewRouter().WithSwagger(spec),
		}

		req, err := http.NewRequest(http.MethodGet, "/", bytes.NewBufferString("test"))
		assert.NoError(t, err)

		err = m.validateRequest(req)
		assert.Error(t, err)
	})

	t.Run("valid request", func(t *testing.T) {
		spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      parameters:
        - name: test
          in: query
          required: true
          schema:
            type: boolean
      responses:
        '200':
          description: "test"
`,
		))

		assert.NoError(t, err)
		m := middleware{
			Context: context.TODO(),
			Router:  openapi3filter.NewRouter().WithSwagger(spec),
		}

		// should return error with no query parameter
		req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)
		err = m.validateRequest(req)
		assert.Error(t, err)

		// shouldn't return error with valid query parameter
		req, err = http.NewRequest(http.MethodGet, "/test?test", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)
		err = m.validateRequest(req)
		assert.Error(t, err)
	})
}

func TestMiddleware_validateResponse(t *testing.T) {
	t.Run("invalid response", func(t *testing.T) {
		spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      responses:
        '200':
          description: "test"
          headers: {}
          content:
            application/json:
              schema:
                type: integer
`,
		))

		assert.NoError(t, err)
		m := middleware{
			Context: context.TODO(),
			Router:  openapi3filter.NewRouter().WithSwagger(spec),
		}

		req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)

		err = m.validateResponse(&httpResponseRecorder{
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Status: 200,
			Body:   bytes.NewBufferString("test"),
		}, req)

		assert.Error(t, err)
	})

	t.Run("valid response", func(t *testing.T) {
		spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      responses:
        '200':
          description: "test"
          headers: {}
          content:
            application/json:
              schema:
                type: integer
`,
		))

		assert.NoError(t, err)
		m := middleware{
			Context: context.TODO(),
			Router:  openapi3filter.NewRouter().WithSwagger(spec),
		}

		req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)

		err = m.validateResponse(&httpResponseRecorder{
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Status: 200,
			Body:   bytes.NewBufferString("200"),
		}, req)

		assert.NoError(t, err)
	})
}

func TestMiddleware_serve(t *testing.T) {
	t.Run("should validate requests", func(t *testing.T) {
		spec := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      parameters:
        - name: test
          in: query
          required: true
          schema:
            type: boolean
      responses:
        '200':
          description: "test"
          headers: {}
          content:
            application/json:
              schema:
                type: integer
`,
		)

		middleware, err := NewMiddleware(&MiddlewareOptions{
			OpenAPISpec: spec,
		})

		assert.NoError(t, err)
		handler := middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

		// test invalid request
		req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, DefaultMiddlewareOptions.InvalidRequestResponse.Status, recorder.Code)

		// test valid request
		req, err = http.NewRequest(http.MethodGet, "/test?test=true", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)
		recorder = httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		assert.NotEqual(t, DefaultMiddlewareOptions.InvalidRequestResponse.Status, recorder.Code)
	})

	t.Run("shouldn't validate responses if opted out", func(t *testing.T) {
		spec := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      responses:
        '200':
          description: "test"
          headers: {}
          content:
            application/json:
              schema:
                type: integer
`,
		)

		middleware, err := NewMiddleware(&MiddlewareOptions{
			OpenAPISpec:      spec,
			ValidateResponse: false,
		})

		assert.NoError(t, err)
		handler := middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
		assert.NoError(t, err)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		assert.NotEqual(t, DefaultMiddlewareOptions.InvalidResponseResponse.Status, recorder.Code)
	})

	t.Run("should validate responses if not opted out", func(t *testing.T) {
		spec := []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths:
  "/test":
    get:
      responses:
        '200':
          description: "test"
          headers: {}
          content:
            application/json:
              schema:
                type: integer
`,
		)

		t.Run("test invalid response", func(t *testing.T) {
			middleware, err := NewMiddleware(&MiddlewareOptions{
				OpenAPISpec:      spec,
				ValidateResponse: true,
			})

			assert.NoError(t, err)
			handler := middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			assert.Equal(t, DefaultMiddlewareOptions.InvalidResponseResponse.Status, recorder.Code)
		})

		t.Run("test valid response", func(t *testing.T) {
			middleware, err := NewMiddleware(&MiddlewareOptions{
				OpenAPISpec:      spec,
				ValidateResponse: true,
			})

			assert.NoError(t, err)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Add("Content-type", "application/json")
				w.Write([]byte("10"))
			}))

			req, err := http.NewRequest(http.MethodGet, "/test", bytes.NewBuffer([]byte{}))
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			assert.NotEqual(t, DefaultMiddlewareOptions.InvalidResponseResponse.Status, recorder.Code)
			assert.Equal(t, []byte("10"), recorder.Body.Bytes())
		})
	})
}
