package ui

import (
	"fmt"
	"strings"
	"time"

	api "SBBuddy/internal/api"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
)

func countChanges(sections []api.Section) int {
	if len(sections) == 0 {
		return 0
	}
	transportSections := 0
	for _, section := range sections {
		if section.Journey != nil {
			transportSections++
		}
	}
	changes := transportSections - 1
	if changes < 0 {
		changes = 0
	}
	return changes
}

func formatChanges(changes int) string {
	if changes == 0 {
		return "Direct"
	}
	return fmt.Sprintf("%d change%s", changes, func() string {
		if changes == 1 {
			return ""
		}
		return "s"
	}())
}

func formatDelay(delay int) string {
	if delay <= 0 {
		return "."
	}
	return fmt.Sprintf("+%d", delay)
}

func formatISOTime(ts string) string {
	if ts == "" {
		return "--:--"
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t.Format("15:04")
		}
	}

	if len(ts) == 5 && strings.Contains(ts, ":") {
		return ts
	}

	if strings.Contains(ts, "T") {
		parts := strings.Split(ts, "T")
		if len(parts) == 2 {
			timePart := parts[1]
			if strings.Contains(timePart, "+") {
				timePart = strings.Split(timePart, "+")[0]
			}
			if strings.Contains(timePart, "Z") {
				timePart = strings.Split(timePart, "Z")[0]
			}
			if strings.Contains(timePart, ":") {
				timeComponents := strings.Split(timePart, ":")
				if len(timeComponents) >= 2 {
					return fmt.Sprintf("%s:%s", timeComponents[0], timeComponents[1])
				}
			}
		}
	}

	if len(ts) >= 5 {
		potential := ts[:5]
		if strings.Contains(potential, ":") {
			return potential
		}
	}

	return "--:--"
}

// parseAPITime attempts to parse timestamps returned by the API. It supports
// both the standard RFC3339 format (with timezone colon) and the variant
// without a colon in the timezone offset (e.g. "+0200").
func parseAPITime(v string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
	}
	var lastErr error
	for _, l := range layouts {
		if t, err := time.Parse(l, v); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}

func durationBetween(start, end string) string {
	t1, err1 := parseAPITime(start)
	t2, err2 := parseAPITime(end)
	if err1 != nil || err2 != nil {
		return "-"
	}
	d := t2.Sub(t1)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%d:%02d", h, m)
}

func renderStationboardTable(sb *api.StationboardResponse) string {
	if sb == nil {
		return ""
	}

	tbl := ltable.New().
		Border(lipgloss.NormalBorder()).
		Headers("Time", "Delay", "Train", "Direction", "Platform")

	for _, e := range sb.Stationboard {
		platform := e.Stop.Platform
		if platform == "" {
			platform = e.Stop.Prognosis.Platform
		}
		if platform == "" {
			platform = "."
		}

		train := fmt.Sprintf("%s%s", e.Category, e.Number)
		direction := fmt.Sprintf("%s → %s", sb.Station.Name, e.To)

		tbl.Row(
			formatISOTime(e.Stop.Departure),
			formatDelay(e.Stop.Delay),
			train,
			direction,
			platform,
		)
	}

	return tbl.String()
}

// RenderStationboardTable exposes stationboard table rendering for CLI usage.
func RenderStationboardTable(sb *api.StationboardResponse) string {
	return renderStationboardTable(sb)
}

// RenderConnectionsTable exposes connection table rendering for CLI usage.
func RenderConnectionsTable(cr *api.ConnectionsResponse) string {
	tbl := buildConnectionsTable(cr)
	return tbl.View()
}

func buildConnectionsTable(cr *api.ConnectionsResponse) table.Model {
	cols := []table.Column{
		{Title: "Departure", Width: 9},
		{Title: "Arrival", Width: 9},
		{Title: "Delay", Width: 6},
		{Title: "Duration", Width: 8},
		{Title: "Changes", Width: 8},
		{Title: "From → To", Width: 25},
	}

	var rows []table.Row
	if cr != nil {
		for _, c := range cr.Connections {
			fromTo := fmt.Sprintf("%s → %s", c.From.Station.Name, c.To.Station.Name)

			duration := c.Duration
			if duration == "" {
				duration = durationBetween(c.From.Departure, c.To.Arrival)
			} else {
				duration = parseDuration(duration)
			}

			changes := countChanges(c.Sections)
			changesStr := formatChanges(changes)

			delayStr := ""
			if c.From.Delay > 0 {
				delayStr = formatDelay(c.From.Delay)
			}

			rows = append(rows, table.Row{
				formatISOTime(c.From.Departure),
				formatISOTime(c.To.Arrival),
				delayStr,
				duration,
				changesStr,
				fromTo,
			})
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+1),
	)
	t.SetStyles(table.DefaultStyles())
	return t
}

// FormatConnectionsTitle builds a multiline title for a connection search.
func FormatConnectionsTitle(from string, via []string, to, info string) string {
	var lines []string
	lines = append(lines, "Connections")
	lines = append(lines, fmt.Sprintf("From: %s", from))
	for i, v := range via {
		lines = append(lines, fmt.Sprintf("Via %d: %s", i+1, v))
	}
	last := fmt.Sprintf("To: %s", to)
	if info != "" {
		lines = append(lines, last)
		lines = append(lines, info)
	} else {
		lines = append(lines, last)
	}
	return strings.Join(lines, "\n")
}

func parseDuration(apiDuration string) string {
	if strings.Contains(apiDuration, "d") {
		parts := strings.Split(apiDuration, "d")
		if len(parts) == 2 {
			timePart := parts[1]
			return parseTimePart(timePart)
		}
	} else {
		return parseTimePart(apiDuration)
	}
	return apiDuration
}

func parseTimePart(timePart string) string {
	parts := strings.Split(timePart, ":")
	if len(parts) >= 2 {
		hours := strings.TrimLeft(parts[0], "0")
		if hours == "" {
			hours = "0"
		}
		minutes := parts[1]
		return fmt.Sprintf("%s:%s", hours, minutes)
	}
	return timePart
}

// parseDateInput converts user input like "24.08.2025" or "2025-08-24" into API format YYYY-MM-DD.
func parseDateInput(v string) (string, error) {
	s := strings.TrimSpace(v)
	if s == "" {
		return time.Now().Format("2006-01-02"), nil
	}
	if t, err := time.Parse("02.01.2006", s); err == nil {
		return t.Format("2006-01-02"), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.Format("2006-01-02"), nil
	}
	return "", fmt.Errorf("invalid date: %s", v)
}

// parseTimeInput validates a time string in HH:mm format. Empty input defaults to current time.
func parseTimeInput(v string) (string, error) {
	s := strings.TrimSpace(v)
	if s == "" {
		return time.Now().Format("15:04"), nil
	}
	if t, err := time.Parse("15:04", s); err == nil {
		return t.Format("15:04"), nil
	}
	return "", fmt.Errorf("invalid time: %s", v)
}

// ParseDateInput is an exported wrapper around parseDateInput.
func ParseDateInput(v string) (string, error) { return parseDateInput(v) }

// ParseTimeInput is an exported wrapper around parseTimeInput.
func ParseTimeInput(v string) (string, error) { return parseTimeInput(v) }

// FormatDateDisplay converts an API date in YYYY-MM-DD format to DD.MM.YYYY for
// displaying titles. It falls back to the input string if parsing fails.
func FormatDateDisplay(apiDate string) string {
	if t, err := time.Parse("2006-01-02", apiDate); err == nil {
		return t.Format("02.01.2006")
	}
	return apiDate
}

// tailUntilClosestArrival returns up to five connections ending with the
// connection whose arrival time is nearest to t.
func tailUntilClosestArrival(conns []api.Connection, t time.Time) []api.Connection {
	if len(conns) <= 5 {
		return conns
	}
	idx := -1
	minDiff := time.Duration(1<<63 - 1)
	for i, c := range conns {
		arr, err := parseAPITime(c.To.Arrival)
		if err != nil {
			continue
		}
		diff := arr.Sub(t)
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			idx = i
		}
	}
	if idx == -1 {
		return conns[:min(len(conns), 5)]
	}
	end := idx + 1
	if end > len(conns) {
		end = len(conns)
	}
	start := end - 5
	if start < 0 {
		start = 0
	}
	return conns[start:end]
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// firstConnectionTime returns the time of the first connection in the list.
// If arrival is true it uses the arrival timestamp, otherwise the departure.
func firstConnectionTime(cr *api.ConnectionsResponse, arrival bool) (time.Time, error) {
	if cr == nil || len(cr.Connections) == 0 {
		return time.Time{}, fmt.Errorf("no connections")
	}
	c := cr.Connections[0]
	ts := c.From.Departure
	if arrival {
		ts = c.To.Arrival
	}
	return parseAPITime(ts)
}

// lastConnectionTime returns the time of the last connection in the list.
// If arrival is true it uses the arrival timestamp, otherwise the departure.
func lastConnectionTime(cr *api.ConnectionsResponse, arrival bool) (time.Time, error) {
	if cr == nil || len(cr.Connections) == 0 {
		return time.Time{}, fmt.Errorf("no connections")
	}
	c := cr.Connections[len(cr.Connections)-1]
	ts := c.From.Departure
	if arrival {
		ts = c.To.Arrival
	}
	return parseAPITime(ts)
}

// firstStationboardTime returns the departure time of the first stationboard entry.
func firstStationboardTime(sb *api.StationboardResponse) (time.Time, error) {
	if sb == nil || len(sb.Stationboard) == 0 {
		return time.Time{}, fmt.Errorf("no stationboard")
	}
	return parseAPITime(sb.Stationboard[0].Stop.Departure)
}

// lastStationboardTime returns the departure time of the last stationboard entry.
func lastStationboardTime(sb *api.StationboardResponse) (time.Time, error) {
	if sb == nil || len(sb.Stationboard) == 0 {
		return time.Time{}, fmt.Errorf("no stationboard")
	}
	last := sb.Stationboard[len(sb.Stationboard)-1]
	return parseAPITime(last.Stop.Departure)
}
