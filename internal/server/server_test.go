package server_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/metriodev/pompiers/internal/adapters/api"
	"github.com/metriodev/pompiers/internal/app"
	"github.com/metriodev/pompiers/internal/pkg/utils"
	"github.com/metriodev/pompiers/internal/server"
)

const (
	mockSigningSecret = "mock-signing-secret"
	mockUser          = "mock-user"
	mockAPIKey        = "mock-api-key"
	mockCloudID       = "mock-cloud-id"
	mockRequestBody   = "mock-request-body"
)

func givenCompassClient(withError bool) *api.CompassClient {
	mockCompassClient := &http.Client{
		Transport: utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if withError {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Internal Server Error"}`)),
				}, nil
			}
			if strings.Contains(req.URL.Path, "/schedules") && !strings.Contains(req.URL.Path, "/on-calls") {
				// Response for GetSchedules
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"values": [{"id": "schedule-1", "name": "Test Schedule"}]}`)),
				}, nil
			} else if strings.Contains(req.URL.Path, "/on-calls") {
				// Response for GetOnCallSchedules
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"onCallParticipants": [{"id": "user-1", "type": "user"}]}`)),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"error": "Not found"}`)),
			}, nil
		}),
	}

	return api.NewCompassClient(mockUser, mockAPIKey, mockCloudID, api.WithHttpClient(mockCompassClient))
}

func givenJiraClient() *api.JiraClient {
	mockJiraClient := &http.Client{
		Transport: utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"accountId": "user-1", "displayName": "Test User", "active": true}`)),
			}, nil
		}),
	}

	return api.NewJiraClient(mockUser, mockAPIKey, api.WithJiraHttpClient(mockJiraClient))
}

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	m.Run()
}

func TestServerEndpoint(t *testing.T) {
	app := app.NewApp(givenCompassClient(false), givenJiraClient())
	port, err := utils.FindFreePort()
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	server := server.NewServer(app, "", port, mockSigningSecret)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(t.Context())

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d", port), io.NopCloser(bytes.NewBuffer([]byte(mockRequestBody))))
	httpReq.Header = utils.GenerateValidSlackHeaders(mockSigningSecret, mockRequestBody)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
	}
	body := string(byteBody)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d with body: %s", http.StatusOK, resp.StatusCode, body)
	}

	if !strings.Contains(body, `{"type":"text","text":"Test Schedule: ","style":{"bold":true}}`) {
		t.Errorf("Expected body to contain 'Test Schedule', got: %s", body)
	}

	if !strings.Contains(body, `{"type":"text","text":"Test Schedule: ","style":{"bold":true}}`) {
		t.Errorf("Expected body to contain 'Test Schedule', got: %s", body)
	}

	if !strings.Contains(body, `{"type":"text","text":"Test User","style":{}}`) {
		t.Errorf("Expected body to contain 'user-1', got: %s", body)
	}
}

func TestErrorMessage(t *testing.T) {
	app := app.NewApp(givenCompassClient(true), givenJiraClient())
	port, err := utils.FindFreePort()
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	server := server.NewServer(app, "", port, mockSigningSecret)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(t.Context())

	httpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d", port), io.NopCloser(bytes.NewBuffer([]byte(mockRequestBody))))
	httpReq.Header = utils.GenerateValidSlackHeaders(mockSigningSecret, mockRequestBody)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	byteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
	}
	body := string(byteBody)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d with body: %s", http.StatusOK, resp.StatusCode, body)
	}

	if !strings.Contains(body, `{"response_type": "ephemeral", "text": "We are having trouble processing this request. Please try again later."}`) {
		t.Errorf("Expected body to erro message.', got: %s", body)
	}
}
