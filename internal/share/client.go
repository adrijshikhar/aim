package share

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	port            int
	httpClient      *http.Client
	baseURLOverride string
}

func NewClient(port int) *Client {
	return &Client{
		port: port,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) RequestLease(ctx context.Context, peerIP, profile, agent string, hours int) (*LeaseResponse, error) {
	baseURL := c.baseURLOverride
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%d", peerIP, c.port)
	}
	reqBody := LeaseRequest{
		RequesterID:    "local-node",
		RequesterIP:    peerIP,
		ProfileName:    profile,
		Agent:          agent,
		RequestedHours: hours,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/share/request", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer returned status %s", resp.Status)
	}

	var leaseResp LeaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&leaseResp); err != nil {
		return nil, err
	}
	return &leaseResp, nil
}

func (c *Client) PollLease(ctx context.Context, peerIP, requestID string) (*LeaseResponse, error) {
	baseURL := c.baseURLOverride
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%d", peerIP, c.port)
	}

	reqURL := fmt.Sprintf("%s/api/share/poll?request_id=%s", baseURL, requestID)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("request %s not found or expired", requestID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer returned status %s", resp.Status)
	}

	var leaseResp LeaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&leaseResp); err != nil {
		return nil, err
	}
	return &leaseResp, nil
}

func (c *Client) CancelLeaseRequest(ctx context.Context, peerIP, requestID string) error {
	baseURL := c.baseURLOverride
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%d", peerIP, c.port)
	}

	reqBody := map[string]string{"request_id": requestID}
	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/share/cancel", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote host returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
