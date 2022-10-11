package api

import (
	"net/http"
	"fmt"
	"context"
	"time"
	"strings"
	"encoding/json"

	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/rest"
	"github.com/go-pkgz/rest/logger"

	"github.com/antibantique/pepe/src/proc"
	"github.com/antibantique/pepe/src/config"
	"github.com/antibantique/pepe/src/source"
)

type Server struct {
	Port             int
	MaxBodySize      int64
	StdOutLogEnbaled bool
	Version          string
	TaskCh           chan *proc.Task
	CommonConf       config.C

	httpServer       *http.Server
}

func (s *Server) Run(ctx context.Context) {
	log.Printf("[INFO] starting http server on port %d", s.Port)

	var err error

	go func() {
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
	msg := make(map[string]string)

	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse post form")
			return
		}

		for k, v := range r.PostForm {
			msg[k] = strings.Join(v, ", ")
		}

	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			rest.SendErrorJSON(w, r, log.Default(), http.StatusInternalServerError, err, "failed to parse json")
			return
		}

	default:
		rest.SendErrorJSON(w, r, log.Default(), http.StatusUnsupportedMediaType, fmt.Errorf("content type error"), "unsupported content type")
		return
	}

	go func() {
		parts := strings.Split(r.RemoteAddr, ":")
		ip := parts[0]

		s.TaskCh <- &proc.Task{
			Src:    &source.S{ Ip: ip, Config: &s.CommonConf },
			RawLog: msg,
		}
	}()

	rest.RenderJSON(w, rest.JSON{"status": "ok"})
}