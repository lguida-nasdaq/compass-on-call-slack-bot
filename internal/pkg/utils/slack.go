package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"
)

// Creates a valid Slack request witht proper signatures
func CreateValidSlackRequest(secret string, body string) *http.Request {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	baseString := "v0:" + timestamp + ":" + body

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", signature)

	return req
}

func GenerateValidSlackHeaders(secret string, body string) http.Header {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	baseString := "v0:" + timestamp + ":" + body

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	h := http.Header{}
	h.Set("X-Slack-Request-Timestamp", timestamp)
	h.Set("X-Slack-Signature", signature)
	h.Set("Content-Type", "application/json")

	return h
}
