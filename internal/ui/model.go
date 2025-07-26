package ui

import (
	"fmt"
	"net/http"
	"time"

	api "SBBuddy/internal/api"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// application states
type appState int

const (
	stateMenu appState = iota
	stateStationInput
	stateStationSuggestions
	stateConnectionInputFrom
	stateConnectionFromSuggestions
	stateConnectionInputTo
	stateConnectionToSuggestions
	stateConnectionInputVia
	stateConnectionViaSuggestions
	stateConnectionReady
	stateDateTimeInput
	stateShowStationboard
	stateShowConnections
	stateLoadingConnections
	stateShowConnectionDetails
	stateShowConnectionQR
)

// Model holds the TUI state
type Model struct {
	state       appState
	returnState appState
	choices     []string
	cursor      int

	stationInput textinput.Model
	fromInput    textinput.Model
	toInput      textinput.Model
	viaInputs    []textinput.Model
	viaStations  []*api.Location
	viaIndex     int

	dateTime     time.Time
	dateField    int
	arrival      bool
	allowArrival bool

	lastSearchTime    time.Time
	lastSearchArrival bool

	// Station suggestions
	suggestions     []api.Location
	selectedStation *api.Location
	fromStation     *api.Location
	toStation       *api.Location

	stationboard *api.StationboardResponse
	connections  *api.ConnectionsResponse
	err          error

	// Loading spinner
	spinner   spinner.Model
	isLoading bool

	api *api.Client

	selectedConnection    *api.Connection // Track which connection is selected
	selectedConnectionIdx int
	detailCursor          int // For navigating within connection details
	qrCode                string

	connTable table.Model

	returnToDetails bool // refresh after fetching connections

	help help.Model
	keys KeyMap
}

// activeKeys returns a copy of the keymap with bindings enabled based on the
// current application state so the help view only displays relevant shortcuts.
func (m *Model) activeKeys() KeyMap {
	k := m.keys

	// Back key isn't available in the main menu
	k.Back.SetEnabled(m.state != stateMenu)

	// QR code generation only in connection details
	k.QR.SetEnabled(m.state == stateShowConnectionDetails)

	// Refresh in views that display results
	switch m.state {
	case stateShowStationboard, stateShowConnections:
		k.Refresh.SetEnabled(true)
		k.DateTime.SetEnabled(true)
	case stateShowConnectionDetails:
		k.Refresh.SetEnabled(true)
		k.DateTime.SetEnabled(false)
	default:
		k.Refresh.SetEnabled(false)
		k.DateTime.SetEnabled(false)
	}

	inDate := m.state == stateDateTimeInput
	inResults := m.state == stateShowConnections || m.state == stateShowStationboard
	k.Now.SetEnabled(inDate)
	k.Left.SetEnabled(inDate || inResults)
	k.Right.SetEnabled(inDate || inResults)

	inputState := m.state == stateConnectionInputFrom || m.state == stateConnectionInputVia || m.state == stateConnectionReady
	k.AddVia.SetEnabled(inputState)
	k.DelVia.SetEnabled(inputState && len(m.viaInputs) > 0)

	k.Modify.SetEnabled(m.state == stateShowConnections)

	return k
}

// prepareDateTime initializes the date/time selection with defaults.
func (m *Model) prepareDateTime(allowArrival, reuse bool) {
	if reuse && !m.lastSearchTime.IsZero() {
		m.dateTime = m.lastSearchTime
		if allowArrival {
			m.arrival = m.lastSearchArrival
		} else {
			m.arrival = false
		}
	} else {
		m.dateTime = time.Now().Truncate(time.Minute)
		m.arrival = false
	}
	m.dateField = 0
	m.allowArrival = allowArrival
}

// adjustDateField increments or decrements the selected part of the date/time.
func (m *Model) adjustDateField(delta int) {
	idx := m.dateField
	if m.allowArrival {
		switch idx {
		case 0:
			m.arrival = !m.arrival
		case 1:
			m.dateTime = m.dateTime.AddDate(0, 0, delta)
		case 2:
			m.dateTime = m.dateTime.AddDate(0, delta, 0)
		case 3:
			m.dateTime = m.dateTime.AddDate(delta, 0, 0)
		case 4:
			m.dateTime = m.dateTime.Add(time.Duration(delta) * time.Hour)
		case 5:
			m.dateTime = m.dateTime.Add(time.Duration(delta) * time.Minute)
		}
	} else {
		switch idx {
		case 0:
			m.dateTime = m.dateTime.AddDate(0, 0, delta)
		case 1:
			m.dateTime = m.dateTime.AddDate(0, delta, 0)
		case 2:
			m.dateTime = m.dateTime.AddDate(delta, 0, 0)
		case 3:
			m.dateTime = m.dateTime.Add(time.Duration(delta) * time.Hour)
		case 4:
			m.dateTime = m.dateTime.Add(time.Duration(delta) * time.Minute)
		}
	}
}

// dateTimeView renders the date/time selection with the active field highlighted.
func (m *Model) dateTimeView() string {
	var parts []string
	if m.allowArrival {
		mode := "Depart"
		if m.arrival {
			mode = "Arrive"
		}
		parts = append(parts, mode)
	}
	parts = append(parts,
		m.dateTime.Format("02"),
		m.dateTime.Format("01"),
		m.dateTime.Format("2006"),
		m.dateTime.Format("15"),
		m.dateTime.Format("04"),
	)
	for i := range parts {
		if i == m.dateField {
			parts[i] = "[" + parts[i] + "]"
		}
	}
	day := m.dateTime.Format("Mon")
	if m.allowArrival {
		return fmt.Sprintf("%s %s %s.%s.%s %s:%s", parts[0], day, parts[1], parts[2], parts[3], parts[4], parts[5])
	}
	return fmt.Sprintf("%s %s.%s.%s %s:%s", day, parts[0], parts[1], parts[2], parts[3], parts[4])
}

func (m *Model) maxDateField() int {
	if m.allowArrival {
		return 5
	}
	return 4
}

// resetConnectionInputs clears all connection-related fields so returning to the
// menu doesn't retain previous queries.
func (m *Model) resetConnectionInputs() {
	m.fromInput.SetValue("")
	m.toInput.SetValue("")
	m.viaInputs = nil
	m.viaStations = nil
	m.viaIndex = 0
	m.fromStation = nil
	m.toStation = nil
}

// InitialModel initializes the Bubble Tea model with optimizations
func InitialModel() *Model {
	input := textinput.New()
	input.Placeholder = "Station name"
	input.Focus()
	input.CharLimit = 50
	input.Width = 40

	from := textinput.New()
	from.Placeholder = "From station"
	from.CharLimit = 50
	from.Width = 40

	to := textinput.New()
	to.Placeholder = "To station"
	to.CharLimit = 50
	to.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	cols := []table.Column{
		{Title: "Departure", Width: 9},
		{Title: "Arrival", Width: 9},
		{Title: "Delay", Width: 6},
		{Title: "Duration", Width: 8},
		{Title: "Changes", Width: 8},
		{Title: "From → To", Width: 25},
	}
	connTbl := table.New(
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
		table.WithHeight(6),
	)
	connTbl.SetStyles(table.DefaultStyles())

	// Create HTTP client with connection pooling and optimized timeouts
	client := &http.Client{
		Timeout: 8 * time.Second, // Slightly reduced timeout
		Transport: &http.Transport{
			MaxIdleConns:        10,               // Allow connection reuse
			MaxIdleConnsPerHost: 5,                // Per host connection reuse
			IdleConnTimeout:     30 * time.Second, // Keep connections alive
			DisableCompression:  false,            // Enable compression
		},
	}

	return &Model{
		state:        stateMenu,
		returnState:  stateMenu,
		choices:      []string{"Lookup timetable", "Find connection", "Random connection"},
		cursor:       0,
		stationInput: input,
		fromInput:    from,
		toInput:      to,
		viaInputs:    nil,
		viaStations:  nil,
		viaIndex:     0,
		dateTime:     time.Now().Truncate(time.Minute),
		allowArrival: false,
		spinner:      s,
		connTable:    connTbl,
		api:          api.NewClient(client),
		help:         help.New(),
		keys:         DefaultKeyMap(),
	}
}
