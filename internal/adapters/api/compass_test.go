package api

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/metriodev/pompiers/internal/config"
	"github.com/metriodev/pompiers/internal/pkg/utils"
)

func TestCompassClient(t *testing.T) {
	// Setup test config
	cfg := config.Config{
		CloudID: "mock-cloud-id",
		User:    "mock-user",
		APIKey:  "mock-api-key",
	}

	// Create a mock HTTP client
	mockClient := &http.Client{
		Transport: utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"values": [{"id": "schedule-1", "name": "Test Schedule"}]}`)),
			}, nil
		}),
	}

	// Create client with the functional option
	client := NewCompassClient(cfg, WithHttpClient(mockClient))

	// Proceed with your tests...
	schedules, err := client.GetSchedules()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}
	if schedules[0].ID != "schedule-1" {
		t.Errorf("expected schedule ID 'schedule-1', got '%s'", schedules[0].ID)
	}
	if schedules[0].Name != "Test Schedule" {
		t.Errorf("expected schedule name 'Test Schedule', got '%s'", schedules[0].Name)
	}
}

// Additional test functions...
