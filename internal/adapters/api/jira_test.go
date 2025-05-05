package api

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/metriodev/pompiers/internal/pkg/utils"
)

func buildMockHttpClient(t *testing.T, res *http.Response) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			user, key, ok := req.BasicAuth()
			if !ok {
				t.Errorf("Missing basic auth credentials")
			}
			if user != mockUser {
				t.Errorf("Expected user '%s', got '%s'", mockUser, user)
			}
			if key != mockApiKey {
				t.Errorf("Expected API key '%s', got '%s'", mockApiKey, key)
			}

			return res, nil
		}),
	}
}

func TestGetUserInfo(t *testing.T) {
	mockResponse := `{
		"accountId": "user-1",
		"accountType": "atlassian",
		"active": true,
		"displayName": "Test User"
	}`

	mockClient := buildMockHttpClient(t, &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(mockResponse)),
	})
	client := NewJiraClient(mockUser, mockApiKey, WithJiraHttpClient(mockClient))

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
}

func TestGetUserInfo_NotFound(t *testing.T) {
	mockClient := buildMockHttpClient(t, &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"errorMessages":["User does not exist"]}`)),
	})
	client := NewJiraClient(mockUser, mockApiKey, WithJiraHttpClient(mockClient))

	_, err := client.GetUserInfo("nonexistent-user")
	if err == nil {
		t.Fatal("expected error for status code 404, got nil")
	}
	if !strings.Contains(err.Error(), "status code 404") {
		t.Errorf("expected error message to contain 'status code 404', got: %v", err)
	}
}

func TestGetUserInfo_InvalidJSON(t *testing.T) {
	// Update the mock client for the invalid JSON test case
	invalidJSON := `{
		"accountId": "user-1",
		"displayName":
	}`
	mockClient := buildMockHttpClient(t, &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(invalidJSON)),
	})
	client := NewJiraClient(mockUser, mockApiKey, WithJiraHttpClient(mockClient))

	_, err := client.GetUserInfo("user-1")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "error decoding response body") {
		t.Errorf("expected error message to contain 'error decoding response body', got: %v", err)
	}
}
