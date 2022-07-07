package api

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxReqSizeHandler(t *testing.T) {
	{
		handler := maxReqSizeHandler(42)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("request: %v", r)
		}))

		rec := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString("424242"))
		require.NoError(t, err)

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode, "set limit and request fits it")
	}
	{
		handler := maxReqSizeHandler(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("request: %v", r)
		}))

		rec := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString("424242"))
		require.NoError(t, err)

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Result().StatusCode, "low limit and big request size")
	}
	{
		handler := maxReqSizeHandler(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("request: %v", r)
		}))

		rec := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString("424242"))
		require.NoError(t, err)

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode, "no limit, status ok")
	}
}