package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/metriodev/pompiers/internal/pkg/utils"
)

const (
	mockSigningSecret = "mock-secret"
)

func setupTest() (http.Handler, *http.Request) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	handler := VerifySlackSignature(mockSigningSecret, testHandler)
	req := utils.CreateValidSlackRequest(mockSigningSecret, "test-body")
	return handler, req
}

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	m.Run()
}

func TestVerifySlackSignature_ValidRequest(t *testing.T) {
	handler, req := setupTest()

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}
}

func TestVerifySlackSignature_InvalidSignature(t *testing.T) {
	handler, req := setupTest()

	// Set an invalid signature
	req.Header.Set("X-Slack-Signature", "invalid-signature")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized, got %v", rr.Code)
	}
}

func TestVerifySlackSignature_ExpiredTimestamp(t *testing.T) {
	handler, req := setupTest()

	// Set an expired timestamp
	expiredTimestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	req.Header.Set("X-Slack-Request-Timestamp", expiredTimestamp)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized for expired timestamp, got %v", rr.Code)
	}
}
