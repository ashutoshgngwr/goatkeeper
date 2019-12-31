package goatkeeper

import (
	"bytes"
	"net/http"
)

// httpResponseRecorder implements http.ResponseWriter and stores response
// data written to it in memory.
type httpResponseRecorder struct {
	Headers http.Header
	Status  int
	Body    *bytes.Buffer
}

func newHTTPResponseRecorder() *httpResponseRecorder {
	return &httpResponseRecorder{
		Headers: http.Header{},
		Body:    bytes.NewBuffer([]byte{}),
	}
}

// empty assignment to conform to the interface.
var _ http.ResponseWriter = &httpResponseRecorder{}

func (w *httpResponseRecorder) Header() http.Header {
	return w.Headers
}

func (w *httpResponseRecorder) Write(body []byte) (int, error) {
	return w.Body.Write(body)
}

func (w *httpResponseRecorder) WriteHeader(status int) {
	w.Status = status
}

func (w *httpResponseRecorder) writeToResponse(writer http.ResponseWriter) error {
	for key, values := range w.Headers {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}

	writer.WriteHeader(w.Status)
	_, err := writer.Write(w.Body.Bytes())
	return err
}
