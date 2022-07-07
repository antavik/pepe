package api

import (
	"net/http"
	"fmt"
	"context"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/rest"
	"github.com/go-pkgz/rest/logger"

	"github.com/antibantique/pepe/src/proc"
)

type Server struct {
	Port             int
	MaxBodySize      int64
	StdOutLogEnbaled bool
	Version          string
	TasksCh          chan *proc.Task

	httpServer       *http.Server
}

func (s *Server) Run(ctx context.Context) {
	log.Printf("[INFO] starting http server on port %d", s.Port)

	var err error

	go func() {
		defer close(s.TasksCh)

		<- ctx.Done()
		if s.httpServer != nil {
			if clsErr := s.httpServer.Close(); clsErr != nil {
				log.Printf("[ERROR] failed to close http server, %v", clsErr)
			}
		}
	}()

	handler := http.NewServeMux()
	handler.HandleFunc("/log", s.log)

	h := rest.Wrap(
		handler,
		rest.Recoverer(log.Default()),
		rest.AppInfo("pepe", "antibantique", s.Version),
		rest.Ping,
		stdoutLogHandler(s.StdOutLogEnbaled, logger.New(logger.Log(log.Default()), logger.Prefix("[INFO]")).Handler),
		maxReqSizeHandler(s.MaxBodySize),
	)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Port),
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	err = s.httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %v", err)
}

func (s *Server) log(w http.ResponseWriter, r *http.Request) {
	var logData map[string][]string

	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse post form")
			return
		}

		logData = r.PostForm
	default:
		rest.SendErrorJSON(w, r, log.Default(), http.StatusUnsupportedMediaType, fmt.Errorf("content type error"), "unsupported content type")
		return
	}

	go func() {
		s.TasksCh <- &proc.Task{
			RemoteAddr: r.RemoteAddr,
			LogData:    logData,
		}
	}()

	rest.RenderJSON(w, rest.JSON{"status": "ok"})
}