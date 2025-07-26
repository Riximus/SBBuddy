package ui

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	api "SBBuddy/internal/api"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdp/qrterminal/v3"
)

func (m *Model) viaNames() []string {
	var names []string
	for _, v := range m.viaStations {
		if v != nil {
			names = append(names, v.Name)
		}
	}
	return names
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle spinner updates
	if m.isLoading {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	// Global keybindings
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	// Handle async messages
	switch msg := msg.(type) {
	case api.ConnectionsMsg:
		m.isLoading = false
		if msg.Connections != nil && m.lastSearchArrival && !m.lastSearchTime.IsZero() {
			msg.Connections.Connections = tailUntilClosestArrival(msg.Connections.Connections, m.lastSearchTime)
		}
		m.connections = msg.Connections
		m.connTable = buildConnectionsTable(msg.Connections)
		m.connTable.Focus()
		m.err = msg.Err
		if m.returnToDetails {
			m.returnToDetails = false
			if m.connections != nil && m.selectedConnectionIdx < len(m.connections.Connections) {
				m.selectedConnection = &m.connections.Connections[m.selectedConnectionIdx]
				m.state = stateShowConnectionDetails
			} else {
				m.state = stateShowConnections
			}
		} else {
			m.cursor = 0
			m.state = stateShowConnections
		}
		return m, nil

	case api.StationboardMsg:
		m.isLoading = false
		m.stationboard = msg.Stationboard
		m.err = msg.Err
		m.state = stateShowStationboard
		return m, nil
	}

	switch m.state {
	case stateMenu:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(keyMsg, m.keys.Down):
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case key.Matches(keyMsg, m.keys.Enter):
				selectedOption := m.cursor
				m.cursor = 0
				if selectedOption == 0 {
					m.stationInput.Focus()
					m.state = stateStationInput
				} else if selectedOption == 1 {
					m.fromInput.Focus()
					m.state = stateConnectionInputFrom
				} else {
					m.isLoading = true
					m.state = stateLoadingConnections
					stations, err := api.RandomStations(0)
					if err != nil {
						m.isLoading = false
						m.err = err
						m.state = stateShowConnections
						return m, nil
					}
					m.fromStation = &api.Location{Name: stations[0]}
					m.toStation = &api.Location{Name: stations[len(stations)-1]}
					via := stations[1 : len(stations)-1]
					m.viaInputs = nil
					m.viaStations = nil
					return m, tea.Batch(m.api.FetchConnectionsCmd(m.fromStation.Name, m.toStation.Name, via), m.spinner.Tick)
				}
			}
		}

	case stateStationInput:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Back) {
			m.stationInput.Blur()
			m.cursor = 0
			m.state = stateMenu
			return m, nil
		}
		m.stationInput, cmd = m.stationInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Enter) {
			query := strings.TrimSpace(m.stationInput.Value())
			if query == "" {
				break
			}
			suggestions, err := m.api.ValidateStation(query)
			if err != nil {
				m.err = fmt.Errorf("failed to search stations: %v", err)
				m.state = stateShowStationboard
				return m, nil
			}

			if len(suggestions) == 0 {
				m.err = fmt.Errorf("no stations found for '%s'", query)
				m.state = stateShowStationboard
				return m, nil
			}

			if exactMatch := api.FindExactMatch(query, suggestions); exactMatch != nil {
				m.selectedStation = exactMatch
				m.stationInput.Blur()
				m.prepareDateTime(false, false)
				m.returnState = stateStationInput
				m.state = stateDateTimeInput
				return m, nil
			}

			m.suggestions = suggestions
			m.cursor = 0
			m.stationInput.Blur()
			m.state = stateStationSuggestions
		}
		return m, cmd

	case stateStationSuggestions:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.stationInput.Focus()
				m.state = stateStationInput
				return m, nil
			case key.Matches(keyMsg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(keyMsg, m.keys.Down):
				if m.cursor < len(m.suggestions)-1 {
					m.cursor++
				}
			case key.Matches(keyMsg, m.keys.Enter):
				m.selectedStation = &m.suggestions[m.cursor]
				m.prepareDateTime(false, false)
				m.returnState = stateStationSuggestions
				m.state = stateDateTimeInput
				return m, nil
			}
		}

	case stateConnectionInputFrom:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.fromInput.Blur()
				m.cursor = 0
				m.resetConnectionInputs()
				m.state = stateMenu
				return m, nil
			case key.Matches(keyMsg, m.keys.AddVia):
				vi := textinput.New()
				vi.Placeholder = "Via station"
				vi.CharLimit = 50
				vi.Width = 40
				vi.Focus()
				m.fromInput.Blur()
				m.viaInputs = append([]textinput.Model{vi}, m.viaInputs...)
				m.viaStations = append([]*api.Location{nil}, m.viaStations...)
				m.viaIndex = 0
				m.state = stateConnectionInputVia
				return m, nil
			case key.Matches(keyMsg, m.keys.Down):
				if len(m.viaInputs) > 0 {
					m.fromInput.Blur()
					m.viaIndex = 0
					m.viaInputs[0].Focus()
					m.state = stateConnectionInputVia
				} else {
					m.fromInput.Blur()
					m.toInput.Focus()
					m.state = stateConnectionInputTo
				}
				return m, nil
			}
		}
		m.fromInput, cmd = m.fromInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Enter) {
			query := strings.TrimSpace(m.fromInput.Value())
			if query == "" {
				break
			}
			suggestions, err := m.api.ValidateStation(query)
			if err != nil {
				m.err = fmt.Errorf("failed to search stations: %v", err)
				m.state = stateShowConnections
				return m, nil
			}

			if len(suggestions) == 0 {
				m.err = fmt.Errorf("no stations found for '%s'", query)
				m.state = stateShowConnections
				return m, nil
			}

			if exactMatch := api.FindExactMatch(query, suggestions); exactMatch != nil {
				m.fromStation = exactMatch
				m.fromInput.Blur()
				if m.toStation != nil {
					m.cursor = 0
					m.state = stateConnectionReady
					return m, nil
				}
				m.toInput.Focus()
				m.state = stateConnectionInputTo
				return m, nil
			}

			m.suggestions = suggestions
			m.cursor = 0
			m.fromInput.Blur()
			m.state = stateConnectionFromSuggestions
		}
		return m, cmd

	case stateConnectionFromSuggestions:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.fromInput.Focus()
				m.state = stateConnectionInputFrom
				return m, nil
			case key.Matches(keyMsg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(keyMsg, m.keys.Down):
				if m.cursor < len(m.suggestions)-1 {
					m.cursor++
				}
			case key.Matches(keyMsg, m.keys.Enter):
				m.fromStation = &m.suggestions[m.cursor]
				if m.toStation != nil {
					m.cursor = 0
					m.state = stateConnectionReady
					return m, nil
				}
				m.toInput.Focus()
				m.cursor = 0
				m.state = stateConnectionInputTo
			}
		}

	case stateConnectionInputTo:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.toInput.Blur()
				m.cursor = 0
				m.fromInput.Focus()
				m.state = stateConnectionInputFrom
				return m, nil
			}
		}

		m.toInput, cmd = m.toInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Enter) {
			query := strings.TrimSpace(m.toInput.Value())
			if query == "" {
				break
			}
			suggestions, err := m.api.ValidateStation(query)
			if err != nil {
				m.err = fmt.Errorf("failed to search stations: %v", err)
				m.state = stateShowConnections
				return m, nil
			}

			if len(suggestions) == 0 {
				m.err = fmt.Errorf("no stations found for '%s'", query)
				m.state = stateShowConnections
				return m, nil
			}

			if exactMatch := api.FindExactMatch(query, suggestions); exactMatch != nil {
				m.toStation = exactMatch
				if m.fromStation != nil {
					m.toInput.Blur()
					m.cursor = len(m.viaInputs) + 2
					m.state = stateConnectionReady
					return m, nil
				}
				m.toInput.Blur()
				m.state = stateConnectionInputFrom
				return m, nil
			}

			m.suggestions = suggestions
			m.cursor = 0
			m.toInput.Blur()
			m.state = stateConnectionToSuggestions
		}
		return m, cmd

	case stateConnectionToSuggestions:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.toInput.Focus()
				m.state = stateConnectionInputTo
				return m, nil
			case key.Matches(keyMsg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(keyMsg, m.keys.Down):
				if m.cursor < len(m.suggestions)-1 {
					m.cursor++
				}
			case key.Matches(keyMsg, m.keys.Enter):
				m.toStation = &m.suggestions[m.cursor]
				if m.fromStation != nil {
					m.toInput.Blur()
					m.cursor = len(m.viaInputs) + 2
					m.state = stateConnectionReady
					return m, nil
				}
				m.toInput.Blur()
				m.state = stateConnectionInputFrom
			}
		}

	case stateConnectionInputVia:
		if len(m.viaInputs) == 0 {
			m.state = stateConnectionInputTo
			return m, nil
		}
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.viaInputs[m.viaIndex].Blur()
				m.toInput.Focus()
				m.state = stateConnectionInputTo
				return m, nil
			case key.Matches(keyMsg, m.keys.Up):
				m.viaInputs[m.viaIndex].Blur()
				if m.viaIndex > 0 {
					m.viaIndex--
					m.viaInputs[m.viaIndex].Focus()
					return m, nil
				}
				m.fromInput.Focus()
				m.state = stateConnectionInputFrom
				return m, nil
			case key.Matches(keyMsg, m.keys.Down):
				m.viaInputs[m.viaIndex].Blur()
				if m.viaIndex < len(m.viaInputs)-1 {
					m.viaIndex++
					m.viaInputs[m.viaIndex].Focus()
					return m, nil
				}
				m.toInput.Focus()
				m.state = stateConnectionInputTo
				return m, nil
			case key.Matches(keyMsg, m.keys.AddVia):
				vi := textinput.New()
				vi.Placeholder = "Via station"
				vi.CharLimit = 50
				vi.Width = 40
				vi.Focus()
				idx := m.viaIndex + 1
				m.viaInputs = append(m.viaInputs[:idx], append([]textinput.Model{vi}, m.viaInputs[idx:]...)...)
				m.viaStations = append(m.viaStations[:idx], append([]*api.Location{nil}, m.viaStations[idx:]...)...)
				m.viaIndex = idx
				return m, nil
			case key.Matches(keyMsg, m.keys.DelVia) && len(m.viaInputs) > 0:
				m.viaInputs = append(m.viaInputs[:m.viaIndex], m.viaInputs[m.viaIndex+1:]...)
				if len(m.viaStations) > m.viaIndex {
					m.viaStations = append(m.viaStations[:m.viaIndex], m.viaStations[m.viaIndex+1:]...)
				}
				if m.viaIndex >= len(m.viaInputs) {
					if len(m.viaInputs) == 0 {
						m.toInput.Focus()
						m.state = stateConnectionInputTo
						m.viaIndex = 0
						return m, nil
					}
					m.viaIndex = len(m.viaInputs) - 1
				}
				m.viaInputs[m.viaIndex].Focus()
				return m, nil
			}
		}
		m.viaInputs[m.viaIndex], cmd = m.viaInputs[m.viaIndex].Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Enter) {
			query := strings.TrimSpace(m.viaInputs[m.viaIndex].Value())
			if query == "" {
				break
			}
			suggestions, err := m.api.ValidateStation(query)
			if err != nil {
				m.err = fmt.Errorf("failed to search stations: %v", err)
				m.state = stateShowConnections
				return m, nil
			}
			if len(suggestions) == 0 {
				m.err = fmt.Errorf("no stations found for '%s'", query)
				m.state = stateShowConnections
				return m, nil
			}
			if exactMatch := api.FindExactMatch(query, suggestions); exactMatch != nil {
				m.viaStations[m.viaIndex] = exactMatch
				m.viaInputs[m.viaIndex].SetValue(exactMatch.Name)
				m.viaInputs[m.viaIndex].Blur()
				m.cursor = m.viaIndex + 1
				m.state = stateConnectionReady
				return m, nil
			}
			m.suggestions = suggestions
			m.cursor = 0
			m.viaInputs[m.viaIndex].Blur()
			m.state = stateConnectionViaSuggestions
		}
		return m, cmd

	case stateConnectionViaSuggestions:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.viaInputs[m.viaIndex].Focus()
				m.state = stateConnectionInputVia
				return m, nil
			case key.Matches(keyMsg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(keyMsg, m.keys.Down):
				if m.cursor < len(m.suggestions)-1 {
					m.cursor++
				}
			case key.Matches(keyMsg, m.keys.Enter):
				sel := &m.suggestions[m.cursor]
				m.viaStations[m.viaIndex] = sel
				m.viaInputs[m.viaIndex].SetValue(sel.Name)
				m.viaInputs[m.viaIndex].Blur()
				m.cursor = m.viaIndex + 1
				m.state = stateConnectionReady
				return m, nil
			}
		}

	case stateConnectionReady:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				m.cursor = 0
				m.resetConnectionInputs()
				m.state = stateMenu
				return m, nil
			case key.Matches(keyMsg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(keyMsg, m.keys.Down):
				max := len(m.viaInputs) + 2
				if m.cursor < max {
					m.cursor++
				}
			case key.Matches(keyMsg, m.keys.Enter):
				switch {
				case m.cursor == 0:
					m.fromInput.Focus()
					m.state = stateConnectionInputFrom
					return m, nil
				case m.cursor >= 1 && m.cursor <= len(m.viaInputs):
					m.viaIndex = m.cursor - 1
					m.viaInputs[m.viaIndex].Focus()
					m.state = stateConnectionInputVia
					return m, nil
				case m.cursor == len(m.viaInputs)+1:
					m.toInput.Focus()
					m.state = stateConnectionInputTo
					return m, nil
				case m.cursor == len(m.viaInputs)+2:
					m.prepareDateTime(true, false)
					m.returnState = stateConnectionReady
					m.state = stateDateTimeInput
					return m, nil
				}
			case key.Matches(keyMsg, m.keys.AddVia):
				if m.cursor == 0 || (m.cursor >= 1 && m.cursor <= len(m.viaInputs)) {
					vi := textinput.New()
					vi.Placeholder = "Via station"
					vi.CharLimit = 50
					vi.Width = 40
					vi.Focus()
					idx := m.cursor
					m.viaInputs = append(m.viaInputs[:idx], append([]textinput.Model{vi}, m.viaInputs[idx:]...)...)
					m.viaStations = append(m.viaStations[:idx], append([]*api.Location{nil}, m.viaStations[idx:]...)...)
					m.viaIndex = idx
					m.state = stateConnectionInputVia
					return m, nil
				}
			case key.Matches(keyMsg, m.keys.DelVia):
				if m.cursor >= 1 && m.cursor <= len(m.viaInputs) {
					idx := m.cursor - 1
					m.viaInputs = append(m.viaInputs[:idx], m.viaInputs[idx+1:]...)
					if len(m.viaStations) > idx {
						m.viaStations = append(m.viaStations[:idx], m.viaStations[idx+1:]...)
					}
					if m.cursor > len(m.viaInputs)+2 {
						m.cursor = len(m.viaInputs) + 2
					} else if m.cursor > len(m.viaInputs)+1 {
						m.cursor = len(m.viaInputs) + 1
					}
				}
			}
		}

	case stateDateTimeInput:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "left":
				if m.dateField > 0 {
					m.dateField--
				}
			case "right":
				if m.dateField < m.maxDateField() {
					m.dateField++
				}
			case "up":
				m.adjustDateField(1)
			case "down":
				m.adjustDateField(-1)
			case "n":
				m.dateTime = time.Now().Truncate(time.Minute)
				m.arrival = false
			default:
				if key.Matches(keyMsg, m.keys.Back) {
					switch m.returnState {
					case stateStationInput:
						m.stationInput.Focus()
					case stateConnectionInputFrom:
						m.fromInput.Focus()
					case stateConnectionInputTo:
						m.toInput.Focus()
					}
					m.state = m.returnState
					return m, nil
				}
				if key.Matches(keyMsg, m.keys.Enter) {
					m.isLoading = true
					m.state = stateLoadingConnections
					m.lastSearchTime = m.dateTime
					m.lastSearchArrival = m.arrival
					date := m.dateTime.Format("2006-01-02")
					tm := m.dateTime.Format("15:04")
					if m.selectedStation != nil {
						return m, tea.Batch(m.api.FetchStationboardAtCmd(m.selectedStation.Name, date, tm), m.spinner.Tick)
					}
					return m, tea.Batch(m.api.FetchConnectionsAtCmd(m.fromStation.Name, m.toStation.Name, m.viaNames(), date, tm, m.arrival), m.spinner.Tick)
				}
			}
		}

	case stateLoadingConnections:
		// Only handle spinner updates and cancellation
		if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, m.keys.Back) {
			m.isLoading = false
			m.cursor = 0
			m.resetConnectionInputs()
			m.state = stateMenu
			return m, nil
		}
		return m, cmd

	case stateShowStationboard:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back) || key.Matches(keyMsg, m.keys.Enter):
				m.state = stateMenu
				m.err = nil
				m.selectedStation = nil
				m.cursor = 0
			case key.Matches(keyMsg, m.keys.Refresh):
				if m.selectedStation != nil {
					m.isLoading = true
					m.state = stateLoadingConnections
					if !m.lastSearchTime.IsZero() {
						date := m.lastSearchTime.Format("2006-01-02")
						tm := m.lastSearchTime.Format("15:04")
						return m, tea.Batch(m.api.FetchStationboardAtCmd(m.selectedStation.Name, date, tm), m.spinner.Tick)
					}
					m.lastSearchTime = time.Now().Truncate(time.Minute)
					m.lastSearchArrival = false
					return m, tea.Batch(m.api.FetchStationboardCmd(m.selectedStation.Name), m.spinner.Tick)
				}
			case key.Matches(keyMsg, m.keys.DateTime):
				if m.selectedStation != nil {
					m.prepareDateTime(false, true)
					m.returnState = stateShowStationboard
					m.state = stateDateTimeInput
					return m, nil
				}
			case key.Matches(keyMsg, m.keys.Left), key.Matches(keyMsg, m.keys.Right):
				if m.selectedStation != nil {
					var t time.Time
					var err error
					if key.Matches(keyMsg, m.keys.Right) {
						t, err = lastStationboardTime(m.stationboard)
					} else {
						if m.lastSearchTime.IsZero() {
							t = time.Now().Add(-10 * time.Minute)
						} else {
							t = m.lastSearchTime.Add(-10 * time.Minute)
						}
					}
					if err == nil {
						m.lastSearchTime = t
						m.isLoading = true
						m.state = stateLoadingConnections
						date := t.Format("2006-01-02")
						tm := t.Format("15:04")
						return m, tea.Batch(m.api.FetchStationboardAtCmd(m.selectedStation.Name, date, tm), m.spinner.Tick)
					}
				}
			}
		}
		return m, nil

	case stateShowConnections:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back) || key.Matches(keyMsg, m.keys.Modify):
				m.state = stateConnectionReady
				m.cursor = 0
			case key.Matches(keyMsg, m.keys.Up), key.Matches(keyMsg, m.keys.Down):
				m.connTable, _ = m.connTable.Update(msg)
				m.cursor = m.connTable.Cursor()
			case key.Matches(keyMsg, m.keys.Enter):
				// Select connection for details
				if m.connections != nil && m.connTable.Cursor() < len(m.connections.Connections) {
					idx := m.connTable.Cursor()
					m.selectedConnection = &m.connections.Connections[idx]
					m.selectedConnectionIdx = idx
					m.detailCursor = 0
					m.state = stateShowConnectionDetails
				}
			case key.Matches(keyMsg, m.keys.Left), key.Matches(keyMsg, m.keys.Right):
				if m.fromStation != nil && m.toStation != nil {
					var t time.Time
					var err error
					if key.Matches(keyMsg, m.keys.Right) {
						t, err = lastConnectionTime(m.connections, m.lastSearchArrival)
					} else {
						if m.lastSearchTime.IsZero() {
							t = time.Now().Add(-10 * time.Minute)
						} else {
							t = m.lastSearchTime.Add(-10 * time.Minute)
						}
					}
					if err == nil {
						m.lastSearchTime = t
						m.isLoading = true
						m.state = stateLoadingConnections
						date := t.Format("2006-01-02")
						tm := t.Format("15:04")
						return m, tea.Batch(m.api.FetchConnectionsAtCmd(m.fromStation.Name, m.toStation.Name, m.viaNames(), date, tm, m.lastSearchArrival), m.spinner.Tick)
					}
				}
			case key.Matches(keyMsg, m.keys.Refresh):
				if m.fromStation != nil && m.toStation != nil {
					m.isLoading = true
					m.state = stateLoadingConnections
					if !m.lastSearchTime.IsZero() {
						date := m.lastSearchTime.Format("2006-01-02")
						tm := m.lastSearchTime.Format("15:04")
						return m, tea.Batch(m.api.FetchConnectionsAtCmd(m.fromStation.Name, m.toStation.Name, m.viaNames(), date, tm, m.lastSearchArrival), m.spinner.Tick)
					}
					m.lastSearchTime = time.Now().Truncate(time.Minute)
					m.lastSearchArrival = false
					return m, tea.Batch(m.api.FetchConnectionsCmd(m.fromStation.Name, m.toStation.Name, m.viaNames()), m.spinner.Tick)
				}
			case key.Matches(keyMsg, m.keys.DateTime):
				if m.fromStation != nil && m.toStation != nil {
					m.prepareDateTime(true, true)
					m.returnState = stateShowConnections
					m.state = stateDateTimeInput
					return m, nil
				}
			}
		}
		return m, nil

	case stateShowConnectionDetails:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(keyMsg, m.keys.Back):
				// back to list
				m.state = stateShowConnections
				m.selectedConnection = nil
			case key.Matches(keyMsg, m.keys.QR):
				// generate QR
				// 1) build deep link:
				from := m.selectedConnection.From.Station
				to := m.selectedConnection.To.Station
				stopsJSON := fmt.Sprintf(
					`[{"value":"%s","type":"ID","label":"%s"},`+
						`{"value":"%s","type":"ID","label":"%s"}]`,
					from.ID, from.Name,
					to.ID, to.Name,
				)

				// 2) Parse the departure timestamp string
				depStr := m.selectedConnection.From.Departure
				depTime, err := parseAPITime(depStr)
				if err != nil {
					// Fallback: use current time (or handle error more gracefully)
					depTime = time.Now()
				}
				depTime = depTime.Local()
				dateStr := depTime.Format("2006-01-02") // YYYY-MM-DD
				timeStr := depTime.Format("15:04")

				// 3) Build and percent-escape each param (including the quotes!)
				stopsParam := url.QueryEscape(stopsJSON)
				dateParam := url.QueryEscape(`"` + dateStr + `"`)
				timeParam := url.QueryEscape(`"` + timeStr + `"`)
				momentParam := url.QueryEscape(`"DEPARTURE"`)
				tripParam := m.selectedConnectionIdx

				deepLink := fmt.Sprintf(
					"https://www.sbb.ch/en?stops=%s&date=%s&time=%s&moment=%s&selected_trip=%d",
					stopsParam, dateParam, timeParam, momentParam, tripParam,
				)

				// 4) Render the QR into a buffer
				var buf strings.Builder
				qrterminal.GenerateWithConfig(deepLink, qrterminal.Config{
					Level:      qrterminal.L,
					Writer:     &buf,
					HalfBlocks: true,
				})

				// 5) Store & switch state to display it
				m.qrCode = buf.String()
				m.state = stateShowConnectionQR
			case key.Matches(keyMsg, m.keys.Refresh):
				if m.fromStation != nil && m.toStation != nil {
					m.isLoading = true
					m.returnToDetails = true
					m.state = stateLoadingConnections
					if !m.lastSearchTime.IsZero() {
						date := m.lastSearchTime.Format("2006-01-02")
						tm := m.lastSearchTime.Format("15:04")
						return m, tea.Batch(m.api.FetchConnectionsAtCmd(m.fromStation.Name, m.toStation.Name, m.viaNames(), date, tm, m.lastSearchArrival), m.spinner.Tick)
					}
					m.lastSearchTime = time.Now().Truncate(time.Minute)
					m.lastSearchArrival = false
					return m, tea.Batch(m.api.FetchConnectionsCmd(m.fromStation.Name, m.toStation.Name, m.viaNames()), m.spinner.Tick)
				}
			}
		}
		return m, nil
	case stateShowConnectionQR:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if key.Matches(keyMsg, m.keys.Back) || key.Matches(keyMsg, m.keys.Enter) {
				m.state = stateShowConnectionDetails
				m.qrCode = ""
			}
		}
		return m, nil
	}

	return m, cmd
}
