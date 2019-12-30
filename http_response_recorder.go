package goatkeeper

import (
	"net/http"
)

type httpResponseRecorder struct {
	headers http.Header
	body    []byte
	status  int
}

var _ http.ResponseWriter = &httpResponseRecorder{}

func (w *httpResponseRecorder) Header() http.Header {
	return w.headers
}

func (w *httpResponseRecorder) Write(body []byte) (int, error) {
	w.body = body
	return 0, nil
}

func (w *httpResponseRecorder) WriteHeader(status int) {
	w.status = status
}

func (w *httpResponseRecorder) writeToHTTPResponseWriter(writer http.ResponseWriter) {
	for key, values := range w.headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	writer.WriteHeader(w.status)

	// TODO add error checking here.
	writer.Write(w.body)
}
