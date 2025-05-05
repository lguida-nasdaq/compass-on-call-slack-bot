package api

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/metriodev/pompiers/internal/pkg/utils"
)

const (
	mockCloudId = "mock-cloud-id"
	mockUser    = "mock-user"
	mockApiKey  = "mock-api-key"
)

func TestCompassClient(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &http.Client{
		Transport: utils.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"values": [{"id": "schedule-1", "name": "Test Schedule"}]}`)),
			}, nil
		}),
	}

	client := NewCompassClient(mockUser, mockApiKey, mockCloudId, WithHttpClient(mockClient))

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
