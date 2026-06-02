// Package tuseme provides the official Go SDK for the Tuseme SMS API.
//
// Usage:
//
//	client := tuseme.NewClient("tk_test_...", "sk_test_...")
//	resp, err := client.Messages.Send(&tuseme.SendRequest{
//	    Content:  "Hello from Tuseme!",
//	    SenderID: "TUSEME-LTD",
//	    Recipients: []tuseme.Recipient{
//	        {MSISDN: "+254712345678", Name: "John Doe"},
//	    },
//	    Type:     "transactional",
//	    Priority: "HIGH",
//	})
package tuseme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://api.tuseme.co.ke/api/v1"
	DefaultTimeout = 30 * time.Second
	DefaultRetries = 3
	retryBackoff   = 500 * time.Millisecond
)

// ── Types ───────────────────────────────────────────────────

type Recipient struct {
	MSISDN string `json:"msisdn"`
	Name   string `json:"name,omitempty"`
}

type SendRequest struct {
	Content      string            `json:"content"`
	SenderID     string            `json:"sender_id,omitempty"`
	Recipients   []Recipient       `json:"recipients,omitempty"`
	GroupIDs     []string          `json:"group_ids,omitempty"`
	ContactIDs   []string          `json:"contact_ids,omitempty"`
	Type         string            `json:"type,omitempty"`
	Priority     string            `json:"priority,omitempty"`
	ScheduledFor string            `json:"scheduled_for,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type SendResponse struct {
	Success          bool    `json:"success"`
	MessageID        string  `json:"message_id"`
	BatchID          string  `json:"batch_id"`
	Status           string  `json:"status"`
	Message          string  `json:"message"`
	EstimatedCost    float64 `json:"estimated_cost"`
	Currency         string  `json:"currency"`
	SelectedProvider string  `json:"selected_provider"`
	RecipientCount   int     `json:"recipient_count"`
	Timestamp        string  `json:"timestamp"`
}

type MessageStatus struct {
	MessageID   string  `json:"message_id"`
	Status      string  `json:"status"`
	Recipient   string  `json:"recipient"`
	SenderID    string  `json:"sender_id"`
	Content     string  `json:"content"`
	Provider    string  `json:"provider"`
	Cost        float64 `json:"cost"`
	Currency    string  `json:"currency"`
	CreatedAt   string  `json:"created_at"`
	DeliveredAt string  `json:"delivered_at"`
}

// ── Error types ────────────────────────────────────────────

type TusemeError struct {
	StatusCode int
	Message    string
	Response   map[string]interface{}
}

func (e *TusemeError) Error() string {
	return fmt.Sprintf("tuseme: %s (status %d)", e.Message, e.StatusCode)
}

// ── Client ─────────────────────────────────────────────────

type Client struct {
	apiKey     string
	apiSecret  string
	baseURL    string
	httpClient *http.Client
	maxRetries int

	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time

	Messages *MessagesResource
}

type Option func(*Client)

func WithBaseURL(url string) Option      { return func(c *Client) { c.baseURL = url } }
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.httpClient.Timeout = d } }
func WithRetries(n int) Option           { return func(c *Client) { c.maxRetries = n } }

func NewClient(apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		maxRetries: DefaultRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Messages = &MessagesResource{client: c}
	return c
}

func (c *Client) IsSandbox() bool    { return len(c.apiKey) > 8 && c.apiKey[:8] == "tk_test_" }
func (c *Client) IsProduction() bool { return len(c.apiKey) > 8 && c.apiKey[:8] == "tk_live_" }

// ── Auth ───────────────────────────────────────────────────

type authResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *Client) authenticate() error {
	body, _ := json.Marshal(map[string]string{
		"api_key":    c.apiKey,
		"api_secret": c.apiSecret,
	})
	resp, err := c.httpClient.Post(c.baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("tuseme: authentication network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &TusemeError{StatusCode: resp.StatusCode, Message: "authentication failed"}
	}

	var result authResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("tuseme: failed to decode auth response: %w", err)
	}

	c.accessToken = result.AccessToken
	c.tokenExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn)*time.Second - 60*time.Second)
	return nil
}

func (c *Client) ensureAuth() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.accessToken == "" || time.Now().After(c.tokenExpiresAt) {
		return c.authenticate()
	}
	return nil
}

// ── Request execution ──────────────────────────────────────

func (c *Client) doRequest(method, path string, payload interface{}, params url.Values) ([]byte, error) {
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		reqURL := c.baseURL + path
		if params != nil {
			reqURL += "?" + params.Encode()
		}

		var bodyReader io.Reader
		if payload != nil {
			b, _ := json.Marshal(payload)
			bodyReader = bytes.NewReader(b)
		}

		req, err := http.NewRequest(method, reqURL, bodyReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "tuseme-go/1.0.0")
		req.Header.Set("Authorization", "Bearer "+c.accessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("tuseme: request failed: %w", err)
			if attempt < c.maxRetries {
				time.Sleep(retryBackoff * time.Duration(math.Pow(2, float64(attempt-1))))
				continue
			}
			return nil, lastErr
		}

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = &TusemeError{StatusCode: resp.StatusCode, Message: string(data)}
			if attempt < c.maxRetries {
				time.Sleep(retryBackoff * time.Duration(math.Pow(2, float64(attempt-1))))
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode >= 400 {
			return nil, &TusemeError{StatusCode: resp.StatusCode, Message: string(data)}
		}

		return data, nil
	}
	return nil, lastErr
}

// ── Messages Resource ──────────────────────────────────────

type MessagesResource struct {
	client *Client
}

func (m *MessagesResource) Send(req *SendRequest) (*SendResponse, error) {
	if req.SenderID == "" {
		req.SenderID = "TUSEME-LTD"
	}
	if req.Type == "" {
		req.Type = "promotional"
	}
	if req.Priority == "" {
		req.Priority = "MEDIUM"
	}

	data, err := m.client.doRequest("POST", "/messages/send", req, nil)
	if err != nil {
		return nil, err
	}

	var resp SendResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("tuseme: failed to decode response: %w", err)
	}
	return &resp, nil
}

func (m *MessagesResource) Get(messageID string) (*MessageStatus, error) {
	data, err := m.client.doRequest("GET", "/messages/"+messageID, nil, nil)
	if err != nil {
		return nil, err
	}

	var status MessageStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("tuseme: failed to decode response: %w", err)
	}
	return &status, nil
}

func (m *MessagesResource) List(page, pageSize int) ([]byte, error) {
	params := url.Values{
		"page":      {fmt.Sprintf("%d", page)},
		"page_size": {fmt.Sprintf("%d", pageSize)},
	}
	return m.client.doRequest("GET", "/messages", nil, params)
}
