package ui

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	api "SBBuddy/internal/api"
)

func loadConn(t *testing.T) *api.Connection {
	var cr api.ConnectionsResponse
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "connections.json"))
	if err != nil {
		t.Fatalf("read connections.json: %v", err)
	}
	if err := json.Unmarshal(data, &cr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cr.Connections) == 0 {
		t.Fatal("no connections")
	}
	return &cr.Connections[0]
}

func TestCountAndFormatChanges(t *testing.T) {
	conn := loadConn(t)
	ch := countChanges(conn.Sections)
	if ch != 1 {
		t.Errorf("expected 1 change, got %d", ch)
	}
	if formatChanges(ch) != "1 change" {
		t.Errorf("formatChanges unexpected: %s", formatChanges(ch))
	}
	if formatChanges(0) != "Direct" {
		t.Errorf("formatChanges direct failed")
	}
}

func TestFormatISOTime(t *testing.T) {
	ts := formatISOTime("2024-01-01T12:30:00Z")
	if ts != "12:30" {
		t.Errorf("unexpected time: %s", ts)
	}
	if formatISOTime("") != "--:--" {
		t.Errorf("empty time failed")
	}
}

func TestDurationHelpers(t *testing.T) {
	d := durationBetween("2024-01-01T12:30:00Z", "2024-01-01T14:20:00Z")
	if d != "1:50" {
		t.Errorf("durationBetween got %s", d)
	}
	if parseDuration("00d01:30:00") != "1:30" {
		t.Errorf("parseDuration failed")
	}
	if parseDuration("01d02:00:00") != "2:00" {
		t.Errorf("parseDuration day handling failed")
	}
	if parseTimePart("01:05:00") != "1:05" {
		t.Errorf("parseTimePart failed")
	}
}

func TestFormatDelay(t *testing.T) {
	if formatDelay(0) != "." {
		t.Errorf("expected dot for no delay")
	}
	if formatDelay(5) != "+5" {
		t.Errorf("delay format wrong")
	}
}

func TestParseDateTimeInput(t *testing.T) {
	if _, err := parseDateInput("today"); err == nil {
		t.Fatalf("expected error for invalid input 'today'")
	}
	d2, err := parseDateInput("24.08.2025")
	if err != nil || d2 != "2025-08-24" {
		t.Errorf("unexpected date %s", d2)
	}
	if _, err := parseDateInput("bad"); err == nil {
		t.Errorf("expected error for bad date")
	}

	t1, err := parseTimeInput("18:34")
	if err != nil || t1 != "18:34" {
		t.Errorf("parseTimeInput failed: %v", err)
	}
	if _, err := parseTimeInput("99:99"); err == nil {
		t.Errorf("expected error for invalid time")
	}
}

func TestRenderTables(t *testing.T) {
	var sb api.StationboardResponse
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "stationboard.json"))
	if err != nil {
		t.Fatalf("read stationboard.json: %v", err)
	}
	if err := json.Unmarshal(data, &sb); err != nil {
		t.Fatalf("unmarshal sb: %v", err)
	}
	out := renderStationboardTable(&sb)
	if !strings.Contains(out, "RE1234") {
		t.Errorf("missing train info")
	}

	conn := loadConn(t)
	out2 := RenderConnectionsTable(&api.ConnectionsResponse{Connections: []api.Connection{*conn}})
	if !strings.Contains(out2, "Chur → Zürich Altstetten") {
		t.Errorf("missing connection direction")
	}
}

func TestRenderStationboardTableNoPlatform(t *testing.T) {
	var sb api.StationboardResponse
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "stationboard.json"))
	if err != nil {
		t.Fatalf("read stationboard.json: %v", err)
	}
	if err := json.Unmarshal(data, &sb); err != nil {
		t.Fatalf("unmarshal sb: %v", err)
	}
	// simulate missing platform
	sb.Stationboard[0].Stop.Platform = ""
	sb.Stationboard[0].Stop.Prognosis.Platform = ""
	out := renderStationboardTable(&sb)
	if !strings.Contains(out, "│.       │") {
		t.Errorf("expected '.' for empty platform")
	}
}

func TestTailUntilClosestArrival(t *testing.T) {
	var cr api.ConnectionsResponse
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "connections.json"))
	if err != nil {
		t.Fatalf("read connections.json: %v", err)
	}
	if err := json.Unmarshal(data, &cr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cr.Connections) < 5 {
		t.Skip("not enough test connections")
	}
	t0, _ := time.Parse(time.RFC3339, cr.Connections[2].To.Arrival)
	res := tailUntilClosestArrival(cr.Connections, t0)
	if len(res) != 5 {
		t.Errorf("expected 5 results, got %d", len(res))
	}
	last := res[len(res)-1]
	if last.To.Arrival != cr.Connections[2].To.Arrival {
		t.Errorf("expected last arrival %s got %s", cr.Connections[2].To.Arrival, last.To.Arrival)
	}
}

func TestFirstLastTimeHelpers(t *testing.T) {
	// Load connection data
	var cr api.ConnectionsResponse
	data, err := ioutil.ReadFile(filepath.Join("..", "..", "testdata", "connections.json"))
	if err != nil {
		t.Fatalf("read connections.json: %v", err)
	}
	if err := json.Unmarshal(data, &cr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cr.Connections) == 0 {
		t.Fatal("no connections")
	}

	first, err := firstConnectionTime(&cr, false)
	if err != nil {
		t.Fatalf("firstConnectionTime error: %v", err)
	}
	last, err := lastConnectionTime(&cr, false)
	if err != nil {
		t.Fatalf("lastConnectionTime error: %v", err)
	}
	if !first.Equal(last) {
		t.Errorf("expected same first and last time, got %v and %v", first, last)
	}

	// Load stationboard data
	var sb api.StationboardResponse
	data, err = ioutil.ReadFile(filepath.Join("..", "..", "testdata", "stationboard.json"))
	if err != nil {
		t.Fatalf("read stationboard.json: %v", err)
	}
	if err := json.Unmarshal(data, &sb); err != nil {
		t.Fatalf("unmarshal sb: %v", err)
	}

	sFirst, err := firstStationboardTime(&sb)
	if err != nil {
		t.Fatalf("firstStationboardTime error: %v", err)
	}
	sLast, err := lastStationboardTime(&sb)
	if err != nil {
		t.Fatalf("lastStationboardTime error: %v", err)
	}
	if !sFirst.Equal(sLast) {
		t.Errorf("expected same first and last stationboard time, got %v and %v", sFirst, sLast)
	}
}
