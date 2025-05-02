package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/metriodev/pompiers/internal/config"
)

func validRequest(cfg config.Config, body string) (*http.Request, string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	baseString := "v0:" + timestamp + ":" + body

	mac := hmac.New(sha256.New, []byte(cfg.SigningSecret))
	mac.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", signature)

	return req, timestamp
}

func TestVerifySlackSignature_ValidRequest(t *testing.T) {
	cfg := config.Config{
		SigningSecret: "test-secret",
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := VerifySlackSignature(cfg, testHandler)

	body := "test-body"
	req, _ := validRequest(cfg, body)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}
}

func TestVerifySlackSignature_InvalidSignature(t *testing.T) {
	cfg := config.Config{
		SigningSecret: "test-secret",
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := VerifySlackSignature(cfg, testHandler)

	body := "test-body"
	req, _ := validRequest(cfg, body)

	// Set an invalid signature
	req.Header.Set("X-Slack-Signature", "invalid-signature")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized, got %v", rr.Code)
	}
}

func TestVerifySlackSignature_ExpiredTimestamp(t *testing.T) {
	cfg := config.Config{
		SigningSecret: "test-secret",
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := VerifySlackSignature(cfg, testHandler)

	body := "test-body"
	req, _ := validRequest(cfg, body)

	// Set an expired timestamp
	expiredTimestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	req.Header.Set("X-Slack-Request-Timestamp", expiredTimestamp)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized for expired timestamp, got %v", rr.Code)
	}
}
