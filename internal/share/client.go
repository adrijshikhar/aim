package share

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
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

// SetBaseURLOverride sets an explicit base URL for testing with httptest.Server.
func (c *Client) SetBaseURLOverride(url string) {
	c.baseURLOverride = url
}

func (c *Client) resolveBaseURL(target string) string {
	if c.baseURLOverride != "" {
		return strings.TrimRight(c.baseURLOverride, "/")
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return strings.TrimRight(target, "/")
	}
	if host, port, err := net.SplitHostPort(target); err == nil {
		return fmt.Sprintf("http://%s", net.JoinHostPort(host, port))
	}
	port := c.port
	if port <= 0 {
		port = 4319
	}
	return fmt.Sprintf("http://%s", net.JoinHostPort(target, strconv.Itoa(port)))
}

func (c *Client) RequestLease(ctx context.Context, peerIP, profile, agent string, hours int) (*LeaseResponse, error) {
	baseURL := c.resolveBaseURL(peerIP)
	requesterID, _ := os.Hostname()
	if requesterID == "" {
		requesterID = "local-node"
	}
	reqBody := LeaseRequest{
		RequesterID:    requesterID,
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		trimmed := strings.TrimSpace(string(body))
		if trimmed != "" {
			return nil, fmt.Errorf("peer returned %s: %s", resp.Status, trimmed)
		}
		return nil, fmt.Errorf("peer returned status %s", resp.Status)
	}

	var leaseResp LeaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&leaseResp); err != nil {
		return nil, err
	}
	return &leaseResp, nil
}

func (c *Client) PollLease(ctx context.Context, peerIP, requestID string) (*LeaseResponse, error) {
	baseURL := c.resolveBaseURL(peerIP)

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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		trimmed := strings.TrimSpace(string(body))
		if trimmed != "" {
			return nil, fmt.Errorf("peer returned %s: %s", resp.Status, trimmed)
		}
		return nil, fmt.Errorf("peer returned status %s", resp.Status)
	}

	var leaseResp LeaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&leaseResp); err != nil {
		return nil, err
	}
	return &leaseResp, nil
}

func (c *Client) CancelLeaseRequest(ctx context.Context, peerIP, requestID string) error {
	baseURL := c.resolveBaseURL(peerIP)

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

type LeaseStatusResponse struct {
	LeaseID string `json:"lease_id"`
	Active  bool   `json:"active"`
	Revoked bool   `json:"revoked"`
	Expired bool   `json:"expired"`
}

func (c *Client) CheckLeaseStatus(ctx context.Context, peerIP, leaseID string) (*LeaseStatusResponse, error) {
	baseURL := c.resolveBaseURL(peerIP)

	reqURL := fmt.Sprintf("%s/api/share/lease/status?lease_id=%s", baseURL, leaseID)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return &LeaseStatusResponse{LeaseID: leaseID, Active: false, Revoked: true, Expired: true}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("host returned status %s", resp.Status)
	}

	var statusResp LeaseStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}
	return &statusResp, nil
}
