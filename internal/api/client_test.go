package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
)

type rewriteTransport struct{ base *url.URL }

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = rt.base.Scheme
	req.URL.Host = rt.base.Host
	return http.DefaultTransport.RoundTrip(req)
}

func loadJSON(t *testing.T, path string, v interface{}) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", path, err)
	}
}

func TestMakeHTTPRequest(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") == "" || r.Header.Get("User-Agent") == "" {
			t.Errorf("missing headers")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := NewClient(&http.Client{Transport: rewriteTransport{u}})

	body, err := client.makeHTTPRequest(context.Background(), baseURL+"/test")
	if err != nil {
		t.Fatalf("makeHTTPRequest error: %v", err)
	}
	if string(body) != "{\"ok\":true}" {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestMakeHTTPRequestStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := NewClient(&http.Client{Transport: rewriteTransport{u}})

	_, err := client.makeHTTPRequest(context.Background(), baseURL+"/fail")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestValidateStationCaching(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "locations.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	stations, err := c.ValidateStation("Chur")
	if err != nil || len(stations) != 5 {
		t.Fatalf("ValidateStation failed: %v", err)
	}
	// call again, should use cache
	stations, err = c.ValidateStation("Chur")
	if err != nil || len(stations) != 5 {
		t.Fatalf("ValidateStation second call: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 server call, got %d", calls)
	}
}

func TestFindExactMatch(t *testing.T) {
	var lr LocationResponse
	loadJSON(t, filepath.Join("..", "..", "testdata", "locations.json"), &lr)

	if loc := FindExactMatch("Chur", lr.Stations); loc == nil || loc.Name != "Chur" {
		t.Errorf("failed exact match for Chur")
	}
	if loc := FindExactMatch("chur", lr.Stations); loc == nil || loc.Name != "Chur" {
		t.Errorf("failed case-insensitive match")
	}
	if loc := FindExactMatch("Foo", lr.Stations); loc != nil {
		t.Errorf("expected nil for unknown station")
	}
}

func TestFetchStationboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "stationboard.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	sb, err := c.FetchStationboard(context.Background(), "Chur")
	if err != nil || sb == nil {
		t.Fatalf("FetchStationboard error: %v", err)
	}
	if len(sb.Stationboard) != 1 {
		t.Errorf("expected 1 entry")
	}
	if sb.Stationboard[0].Name != "RE" {
		t.Errorf("unexpected train name: %s", sb.Stationboard[0].Name)
	}
}

func TestFetchStationboardAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("datetime") != "2025-01-01 10:00" {
			t.Errorf("unexpected datetime %s", r.URL.Query().Get("datetime"))
		}
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "stationboard.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	sb, err := c.FetchStationboardAt(context.Background(), "Chur", "2025-01-01", "10:00")
	if err != nil || sb == nil {
		t.Fatalf("FetchStationboardAt error: %v", err)
	}
}

func TestFetchConnections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "connections.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	cr, err := c.FetchConnections(context.Background(), "Chur", "Zürich Altstetten", nil)
	if err != nil || cr == nil {
		t.Fatalf("FetchConnections error: %v", err)
	}
	if len(cr.Connections) != 1 {
		t.Errorf("expected 1 connection")
	}
	if cr.Connections[0].From.Station.Name != "Chur" {
		t.Errorf("unexpected from station")
	}
}

func TestFetchConnectionsVia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()["via[]"]
		if len(v) != 2 || v[0] != "Bern" || v[1] != "Olten" {
			t.Errorf("unexpected via params %v", v)
		}
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "connections.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	cr, err := c.FetchConnections(context.Background(), "Chur", "Basel", []string{"Bern", "Olten"})
	if err != nil || cr == nil {
		t.Fatalf("FetchConnections via error: %v", err)
	}
}

func TestFetchConnectionsAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("date") != "2025-01-01" || q.Get("time") != "10:00" || q.Get("isArrivalTime") != "1" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "connections.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	cr, err := c.FetchConnectionsAt(context.Background(), "Chur", "Zürich Altstetten", nil, "2025-01-01", "10:00", true)
	if err != nil || cr == nil {
		t.Fatalf("FetchConnectionsAt error: %v", err)
	}
}

func TestFetchConnectionsAtVia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		via := q["via[]"]
		if len(via) != 1 || via[0] != "Bern" {
			t.Errorf("unexpected via params %v", via)
		}
		http.ServeFile(w, r, filepath.Join("..", "..", "testdata", "connections.json"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	c := NewClient(&http.Client{Transport: rewriteTransport{u}})

	cr, err := c.FetchConnectionsAt(context.Background(), "Chur", "Basel", []string{"Bern"}, "2025-01-01", "10:00", true)
	if err != nil || cr == nil {
		t.Fatalf("FetchConnectionsAt via error: %v", err)
	}
}

func TestValidateStationShortQuery(t *testing.T) {
	c := NewClient(nil)
	if _, err := c.ValidateStation("A"); err == nil {
		t.Fatal("expected error for short query")
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 {
		t.Errorf("min failed")
	}
	if min(5, 3) != 3 {
		t.Errorf("min failed")
	}
}
