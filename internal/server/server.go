package server

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/metriodev/pompiers/internal/adapters/api"
	"github.com/metriodev/pompiers/internal/app"
	"github.com/metriodev/pompiers/internal/config"
	"github.com/metriodev/pompiers/internal/middleware"
	"github.com/metriodev/pompiers/internal/pkg/slackmsg"
)

type Server struct {
	Config config.Config
	App    *app.App
}

func NewServer(cfg config.Config, apiClient *api.CompassClient, adminClient *api.JiraClient) *Server {
	return &Server{
		Config: cfg,
		App:    app.NewApp(cfg, *apiClient, *adminClient),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		currentSchedule, err := s.App.GetCurrentOnCallSchedule()
		if err != nil {
			if appErr, ok := err.(app.AppError); ok {
				http.Error(w, appErr.Error(), appErr.HttpCode)
			} else {
				slog.Error("Error fetching current on-call schedule", "error", err)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response_type": "ephemeral", "text": "We are having trouble to process this request. Please try again later."   }`))
				return
			}
		}

		response, err := slackmsg.ToSlackMessage(currentSchedule)
		if err != nil {
			slog.Error("Error converting to Slack message", "error", err, "schedule", currentSchedule)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"response_type": "ephemeral", "text": "We are having trouble to process this request. Please try again later."   }`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(response)
	})

	// Wrap the handler with the Slack signature verification middleware
	log.Println("Server is running on port 8080...")
	return http.ListenAndServe(":8080", middleware.VerifySlackSignature(s.Config, mux))
}
