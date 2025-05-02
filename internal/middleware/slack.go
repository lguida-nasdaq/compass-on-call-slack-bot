package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/metriodev/pompiers/internal/config"
	"github.com/slack-go/slack"
)

var (
	maxTimeElapsed = 5 * time.Minute
)

func VerifySlackSignature(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.SigningSecret == "" {
			http.Error(w, "Slack signing secret not configured", http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading rsequest body", "error", err, "body", string(body))
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		sv, err := slack.NewSecretsVerifier(r.Header, cfg.SigningSecret)
		if err != nil {
			slog.Error("Error creating secret verifier", "body", string(body), "headers", r.Header, "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if _, err := sv.Write(body); err != nil {
			slog.Error("Error writing body to secrets verifier", "body", string(body), "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Restore body for downstream handlers
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		next.ServeHTTP(w, r)
	})
}
