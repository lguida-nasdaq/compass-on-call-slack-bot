package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/metriodev/pompiers/internal/app"
	"github.com/metriodev/pompiers/internal/middleware"
	"github.com/metriodev/pompiers/internal/pkg/slackmsg"
)

const (
	errMsg = "We are having trouble processing this request. Please try again later."
)

type Server struct {
	host               string
	port               int
	slackSigningSecret string
	app                *app.App
	httpserver         *http.Server
}

func NewServer(app *app.App, host string, port int, slackSigningSecret string) *Server {
	if port == 0 {
		port = 8080
	}
	return &Server{
		host:               host,
		port:               port,
		app:                app,
		slackSigningSecret: slackSigningSecret,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		currentSchedule, err := s.app.GetCurrentOnCallSchedule()
		if err != nil {
			if appErr, ok := err.(app.AppError); ok {
				http.Error(w, appErr.Error(), appErr.HttpCode)
			} else {
				slog.Error("Error fetching current on-call schedule", "error", err)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{"response_type": "ephemeral", "text": "%s"}`, errMsg)))
				return
			}
		}

		response, err := slackmsg.ToSlackMessage(currentSchedule)
		if err != nil {
			slog.Error("Error converting to Slack message", "error", err, "schedule", currentSchedule)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response_type":"ephemeral","text":"We are having trouble to process this request. Please try again later."}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})

	s.httpserver = &http.Server{
		Handler: middleware.VerifySlackSignature(s.slackSigningSecret, mux),
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("net.Listen: %v", err)
	}
	go func(l net.Listener) {
		slog.Info(fmt.Sprintf("Server started on %s", l.Addr().String()))
		if err := s.httpserver.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}(listener)

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpserver.Shutdown(ctx)
}
