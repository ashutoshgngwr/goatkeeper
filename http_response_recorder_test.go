package goatkeeper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpResponseRecorder(t *testing.T) {
	t.Run("test recorded data", func(t *testing.T) {
		recorder := newHTTPResponseRecorder()
		recorder.Header().Add("test", "value")
		recorder.WriteHeader(http.StatusAccepted)
		recorder.Write([]byte("hello, world"))

		assert.Equal(t, "value", recorder.Headers.Get("test"))
		assert.Equal(t, recorder.Status, http.StatusAccepted)
		assert.Equal(t, []byte("hello, world"), recorder.Body.Bytes())

		anotherRecorderWhichCouldHaveBeenUsed := httptest.NewRecorder()
		recorder.writeToResponse(anotherRecorderWhichCouldHaveBeenUsed)
		assert.Equal(t, recorder.Headers, anotherRecorderWhichCouldHaveBeenUsed.HeaderMap)
		assert.Equal(t, recorder.Status, anotherRecorderWhichCouldHaveBeenUsed.Code)
		assert.Equal(t, recorder.Body.Bytes(), anotherRecorderWhichCouldHaveBeenUsed.Body.Bytes())
	})
}
