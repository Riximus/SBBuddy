package ui

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	api "SBBuddy/internal/api"
)

func TestRenderConnectionDetails(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "connections.json"))
	if err != nil {
		t.Fatalf("read connections.json: %v", err)
	}
	var cr api.ConnectionsResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	conn := &cr.Connections[0]

	out := renderConnectionDetails(conn)
	dep := formatISOTime(conn.From.Departure)
	arr := formatISOTime(conn.To.Arrival)
	dur := parseDuration(conn.Duration)
	checks := []string{
		"Chur → Zürich Altstetten",
		dep,
		arr,
		dur,
		"Platform 1",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestRenderConnectionDetailsNoPlatform(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "connections.json"))
	if err != nil {
		t.Fatalf("read connections.json: %v", err)
	}
	var cr api.ConnectionsResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	conn := &cr.Connections[0]
	// Remove platform info
	for i := range conn.Sections {
		conn.Sections[i].Departure.Platform = ""
		conn.Sections[i].Departure.Prognosis.Platform = ""
		conn.Sections[i].Arrival.Platform = ""
		conn.Sections[i].Arrival.Prognosis.Platform = ""
	}
	out := renderConnectionDetails(conn)
	if strings.Contains(out, "Platform ") {
		t.Errorf("expected no platform label when platform missing")
	}
}
