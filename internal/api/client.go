package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	baseURL         = "https://transport.opendata.ch/v1"
	endpointStation = "/stationboard"
	endpointConn    = "/connections"
	endpointLoc     = "/locations"
)

type Client struct {
	httpClient   *http.Client
	stationCache map[string][]Location
	cacheMutex   sync.RWMutex
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	return &Client{
		httpClient:   httpClient,
		stationCache: make(map[string][]Location),
	}
}

func (c *Client) makeHTTPRequest(ctx context.Context, requestURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("User-Agent", "SwissTransportTUI/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("received empty response from API")
	}

	trimmed := strings.TrimSpace(string(body))
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return nil, fmt.Errorf("received non-JSON response: %s", string(body)[:min(100, len(body))])
	}

	return body, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) ValidateStation(query string) ([]Location, error) {
	if len(strings.TrimSpace(query)) < 2 {
		return nil, fmt.Errorf("query too short")
	}

	normalized := strings.ToLower(strings.TrimSpace(query))

	c.cacheMutex.RLock()
	if cached, ok := c.stationCache[normalized]; ok {
		c.cacheMutex.RUnlock()
		return cached, nil
	}
	c.cacheMutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	encodedQuery := url.QueryEscape(strings.TrimSpace(query))
	requestURL := fmt.Sprintf("%s%s?query=%s&type=station", baseURL, endpointLoc, encodedQuery)

	body, err := c.makeHTTPRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	var result LocationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	c.cacheMutex.Lock()
	c.stationCache[normalized] = result.Stations
	if len(c.stationCache) > 50 {
		for k := range c.stationCache {
			delete(c.stationCache, k)
			if len(c.stationCache) <= 40 {
				break
			}
		}
	}
	c.cacheMutex.Unlock()

	return result.Stations, nil
}

func FindExactMatch(query string, suggestions []Location) *Location {
	queryLower := strings.ToLower(strings.TrimSpace(query))
	for _, station := range suggestions {
		if strings.ToLower(station.Name) == queryLower {
			return &station
		}
	}
	return nil
}

type StationboardMsg struct {
	Stationboard *StationboardResponse
	Err          error
}

type ConnectionsMsg struct {
	Connections *ConnectionsResponse
	Err         error
}

func (c *Client) FetchStationboard(ctx context.Context, station string) (*StationboardResponse, error) {
	encoded := url.QueryEscape(strings.TrimSpace(station))
	requestURL := fmt.Sprintf("%s%s?station=%s&limit=10", baseURL, endpointStation, encoded)

	body, err := c.makeHTTPRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	var sb StationboardResponse
	if err := json.Unmarshal(body, &sb); err != nil {
		return nil, fmt.Errorf("failed to parse stationboard JSON: %v", err)
	}
	return &sb, nil
}

func (c *Client) FetchStationboardAt(ctx context.Context, station, date, timeStr string) (*StationboardResponse, error) {
	encoded := url.QueryEscape(strings.TrimSpace(station))
	requestURL := fmt.Sprintf("%s%s?station=%s&limit=10", baseURL, endpointStation, encoded)
	if date != "" && timeStr != "" {
		dt := url.QueryEscape(fmt.Sprintf("%s %s", date, timeStr))
		requestURL += "&datetime=" + dt
	}

	body, err := c.makeHTTPRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	var sb StationboardResponse
	if err := json.Unmarshal(body, &sb); err != nil {
		return nil, fmt.Errorf("failed to parse stationboard JSON: %v", err)
	}
	return &sb, nil
}

func (c *Client) FetchConnections(ctx context.Context, from, to string, via []string) (*ConnectionsResponse, error) {
	encodedFrom := url.QueryEscape(strings.TrimSpace(from))
	encodedTo := url.QueryEscape(strings.TrimSpace(to))
	requestURL := fmt.Sprintf("%s%s?from=%s&to=%s&limit=5", baseURL, endpointConn, encodedFrom, encodedTo)
	for _, v := range via {
		requestURL += "&via[]=" + url.QueryEscape(strings.TrimSpace(v))
	}

	body, err := c.makeHTTPRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	var cr ConnectionsResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("failed to parse connections JSON: %v", err)
	}
	return &cr, nil
}

func (c *Client) FetchConnectionsAt(ctx context.Context, from, to string, via []string, date, timeStr string, arrival bool) (*ConnectionsResponse, error) {
	encodedFrom := url.QueryEscape(strings.TrimSpace(from))
	encodedTo := url.QueryEscape(strings.TrimSpace(to))
	requestURL := fmt.Sprintf("%s%s?from=%s&to=%s&limit=5", baseURL, endpointConn, encodedFrom, encodedTo)
	for _, v := range via {
		requestURL += "&via[]=" + url.QueryEscape(strings.TrimSpace(v))
	}
	if date != "" {
		requestURL += "&date=" + url.QueryEscape(date)
	}
	if timeStr != "" {
		requestURL += "&time=" + url.QueryEscape(timeStr)
	}
	if arrival {
		requestURL += "&isArrivalTime=1"
	}

	body, err := c.makeHTTPRequest(ctx, requestURL)
	if err != nil {
		return nil, err
	}

	var cr ConnectionsResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("failed to parse connections JSON: %v", err)
	}
	return &cr, nil
}

func (c *Client) FetchStationboardCmd(station string) tea.Cmd {
	return func() tea.Msg {
		sb, err := c.FetchStationboard(context.Background(), station)
		return StationboardMsg{sb, err}
	}
}

func (c *Client) FetchStationboardAtCmd(station, date, timeStr string) tea.Cmd {
	return func() tea.Msg {
		sb, err := c.FetchStationboardAt(context.Background(), station, date, timeStr)
		return StationboardMsg{sb, err}
	}
}

func (c *Client) FetchConnectionsCmd(from, to string, via []string) tea.Cmd {
	return func() tea.Msg {
		cr, err := c.FetchConnections(context.Background(), from, to, via)
		return ConnectionsMsg{cr, err}
	}
}

func (c *Client) FetchConnectionsAtCmd(from, to string, via []string, date, timeStr string, arrival bool) tea.Cmd {
	return func() tea.Msg {
		cr, err := c.FetchConnectionsAt(context.Background(), from, to, via, date, timeStr, arrival)
		return ConnectionsMsg{cr, err}
	}
}
