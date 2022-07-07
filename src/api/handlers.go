package api

import (
	"net/http"
	"strings"

	log "github.com/go-pkgz/lgr"
)

func passHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func maxReqSizeHandler(maxSize int64) func(next http.Handler) http.Handler {
	if maxSize <= 0 {
		return passHandler
	}

	log.Printf("[DEBUG] request size limited to %d", maxSize)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxSize {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func stdoutLogHandler(enable bool, logHandler func(next http.Handler) http.Handler) func(next http.Handler) http.Handler {
	if !enable {
		return passHandler
	}

	log.Printf("[DEBUG] stdout logging enabled")

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/ping") {
				next.ServeHTTP(w, r)
				return
			}
			logHandler(next).ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}