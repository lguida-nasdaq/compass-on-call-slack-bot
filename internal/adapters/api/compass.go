package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/metriodev/pompiers/internal/config"
)

const (
	baseUrl = "https://api.atlassian.com/compass/cloud"
)

// ClientOption allows for functional options to configure the CompassClient
type ClientOption func(*CompassClient)

// WithHttpClient sets a custom HTTP client
func WithHttpClient(client *http.Client) ClientOption {
	return func(c *CompassClient) {
		c.client = client
	}
}

type CompassClient struct {
	config config.Config
	client *http.Client
}

func NewCompassClient(cfg config.Config, opts ...ClientOption) *CompassClient {
	client := &CompassClient{
		config: cfg,
		client: &http.Client{},
	}

	// Apply the functional options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

type compassApiRequest struct {
	Endpoint string
	Method   string
	Body     io.Reader
}

type scheduleAPIResponse struct {
	Values []Schedule `json:"values"`
}

func (c *CompassClient) doRequest(req compassApiRequest) (*http.Response, error) {
	endpoint, err := url.JoinPath(baseUrl, c.config.CloudID, "/ops/v1")
	if err != nil {
		return nil, fmt.Errorf("error joining base URL: %w", err)
	}

	endpoint, err = url.JoinPath(endpoint, req.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("error joining API URL: %w", err)
	}

	httpReq, err := http.NewRequest(req.Method, endpoint, req.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.SetBasicAuth(c.config.User, c.config.APIKey)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	log.Printf("Requesting URL: %s", httpReq.URL.String())
	res, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return res, nil
}

func (c *CompassClient) GetSchedules() ([]Schedule, error) {
	req := compassApiRequest{
		Endpoint: "schedules",
		Method:   "GET",
		Body:     nil,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error: received status code %d, body: %s", res.StatusCode, body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var response scheduleAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Response body: %s", body)
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return response.Values, nil
}

func (c *CompassClient) GetOnCallSchedules(scheduleID string) (*OnCallResponse, error) {
	req := compassApiRequest{
		Endpoint: fmt.Sprintf("schedules/%s/on-calls", scheduleID),
		Method:   "GET",
		Body:     nil,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error: received status code %d, body: %s", res.StatusCode, body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var onCallResponse OnCallResponse

	if err := json.Unmarshal(body, &onCallResponse); err != nil {
		log.Printf("Response body: %s", body)
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return &onCallResponse, nil
}
