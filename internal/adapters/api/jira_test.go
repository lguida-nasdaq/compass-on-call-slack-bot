package api

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/metriodev/pompiers/internal/config"
	"github.com/metriodev/pompiers/internal/pkg/utils"
)

func TestJiraClient(t *testing.T) {
	// Setup test config
	cfg := config.Config{
		User:   "mock-user",
		APIKey: "mock-api-key",
	}

	// Create a mock HTTP client with the test response for GetUserInfo
	mockResponse := `{
		"accountId": "user-1",
		"accountType": "atlassian",
		"active": true,
		"displayName": "Test User"
	}`

	mockClient := &http.Client{
		Transport: utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(mockResponse)),
			}, nil
		}),
	}

	// Create client with the functional option
	client := NewJiraClient(cfg, WithJiraHttpClient(mockClient))

	// Test your client methods...
	user, err := client.GetUserInfo("user-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.AccountID != "user-1" {
		t.Errorf("expected account ID 'user-1', got '%s'", user.AccountID)
	}
	if user.DisplayName != "Test User" {
		t.Errorf("expected display name 'Test User', got '%s'", user.DisplayName)
	}
	if !user.Active {
		t.Errorf("expected user to be active")
	}

	// Update the mock client for the next test case (nonexistent user)
	mockClient.Transport = utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"errorMessages":["User does not exist"]}`)),
		}, nil
	})

	_, err = client.GetUserInfo("nonexistent-user")
	if err == nil {
		t.Fatal("expected error for status code 404, got nil")
	}
	if !strings.Contains(err.Error(), "status code 404") {
		t.Errorf("expected error message to contain 'status code 404', got: %v", err)
	}

	// Update the mock client for the invalid JSON test case
	invalidJSON := `{
		"accountId": "user-1",
		"displayName":
	}`
	mockClient.Transport = utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(invalidJSON)),
		}, nil
	})

	_, err = client.GetUserInfo("user-1")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "error decoding response body") {
		t.Errorf("expected error message to contain 'error decoding response body', got: %v", err)
	}
}
