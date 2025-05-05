package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

const (
	apiBaseUrl = "https://nasdaq-metrio.atlassian.net/rest/api/3"
)

// ClientOption allows for functional options to configure the JiraClient
type JiraClientOption func(*JiraClient)

// WithHttpClient sets a custom HTTP client for JiraClient
func WithJiraHttpClient(client *http.Client) JiraClientOption {
	return func(j *JiraClient) {
		j.client = client
	}
}

type JiraClient struct {
	user   string
	apiKey string
	client *http.Client
}

// Replace current constructor with this one
func NewJiraClient(user, apiKey string, opts ...JiraClientOption) *JiraClient {
	client := &JiraClient{
		user:   user,
		apiKey: apiKey,
		client: &http.Client{},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type User struct {
	AccountID   string `json:"accountId"`
	AccountType string `json:"accountType"`
	Active      bool   `json:"active"`
	DisplayName string `json:"displayName"`
}

type jiraApiRequest struct {
	Endpoint string
	Method   string
	Body     io.Reader
	Query    url.Values
}

type apiErrorResponse struct {
	Type     string   `json:"type"`
	ApiError apiError `json:"error"`
}

type apiError struct {
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Data    string `json:"data"`
}

func (c *JiraClient) doRequest(req jiraApiRequest) (*http.Response, error) {
	apiUrl, err := url.JoinPath(apiBaseUrl, req.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("error joining API URL: %w", err)
	}

	httpReq, err := http.NewRequest(req.Method, apiUrl, req.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.URL.RawQuery = req.Query.Encode()
	httpReq.SetBasicAuth(c.user, c.apiKey)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return res, nil
}

func (c *JiraClient) GetUserInfo(accountID string) (*User, error) {
	slog.Info("Fetching user info", slog.String("accountID", accountID))
	req := jiraApiRequest{
		Endpoint: "user",
		Query:    url.Values{"accountId": {accountID}},
		Method:   "GET",
		Body:     nil,
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var apiErr apiErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("error decoding error response: %w", err)
		}
		slog.Error(
			"Error fetching user info",
			slog.String("httpCope", res.Status),
			slog.String("accountID", accountID),
			slog.String("errorType", apiErr.Type),
			slog.Group("error",
				slog.String("message", apiErr.ApiError.Message),
				slog.String("detail", apiErr.ApiError.Detail),
				slog.String("data", apiErr.ApiError.Data),
			),
		)
		return nil, fmt.Errorf("error: received status code %d, message: %s", res.StatusCode, apiErr.ApiError.Message)
	}

	var user User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return &user, nil
}
