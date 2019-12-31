package goatkeeper

import (
	"net/http"
)

// httpResponseRecorder implements http.ResponseWriter and stores response
// data written to it in memory.
type httpResponseRecorder struct {
	MiddlewareResponse // already has everything we need.
}

// empty assignment to conform to the interface.
var _ http.ResponseWriter = &httpResponseRecorder{}

func (w *httpResponseRecorder) Header() http.Header {
	return w.Headers
}

func (w *httpResponseRecorder) Write(body []byte) (int, error) {
	w.Body = body
	return 0, nil
}

func (w *httpResponseRecorder) WriteHeader(status int) {
	w.Status = status
}
