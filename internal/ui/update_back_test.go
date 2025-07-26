package ui

import (
	"testing"

	api "SBBuddy/internal/api"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestBackFromDateTimeReturns(t *testing.T) {
	m := InitialModel()
	// simulate connections screen
	m.state = stateShowConnections
	m.fromStation = &api.Location{Name: "A"}
	m.toStation = &api.Location{Name: "B"}
	m.returnState = m.state
	m.state = stateDateTimeInput

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm := newModel.(*Model)
	if nm.state != stateShowConnections {
		t.Fatalf("expected %v got %v", stateShowConnections, nm.state)
	}
}

func TestBackFocusStationInput(t *testing.T) {
	m := InitialModel()
	m.state = stateStationInput
	m.stationInput.Blur()
	m.returnState = m.state
	m.state = stateDateTimeInput

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm := newModel.(*Model)
	if nm.state != stateStationInput {
		t.Fatalf("expected %v got %v", stateStationInput, nm.state)
	}
	if !nm.stationInput.Focused() {
		t.Errorf("station input should be focused")
	}
}

func TestBackFromShowConnectionsReturnsToReady(t *testing.T) {
	m := InitialModel()
	m.state = stateShowConnections
	m.fromStation = &api.Location{Name: "A"}
	m.toStation = &api.Location{Name: "B"}
	m.fromInput.SetValue("foo")
	m.toInput.SetValue("bar")
	vi := textinput.New()
	vi.SetValue("via")
	m.viaInputs = []textinput.Model{vi}
	m.viaStations = []*api.Location{{Name: "V"}}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm := newModel.(*Model)
	if nm.state != stateConnectionReady {
		t.Fatalf("expected %v got %v", stateConnectionReady, nm.state)
	}
	if nm.fromInput.Value() != "foo" || nm.toInput.Value() != "bar" || len(nm.viaInputs) != 1 {
		t.Errorf("inputs changed unexpectedly")
	}
}

func TestModifyKeyFromShowConnections(t *testing.T) {
	m := InitialModel()
	m.state = stateShowConnections
	m.fromStation = &api.Location{Name: "A"}
	m.toStation = &api.Location{Name: "B"}
	newModel, _ := m.Update(tea.KeyMsg{Runes: []rune{'c'}, Type: tea.KeyRunes})
	nm := newModel.(*Model)
	if nm.state != stateConnectionReady {
		t.Fatalf("expected %v got %v", stateConnectionReady, nm.state)
	}
}

func TestViaSuggestionSelectKeepsPosition(t *testing.T) {
	m := InitialModel()
	vi := textinput.New()
	vi.SetValue("Ol")
	m.viaInputs = []textinput.Model{vi}
	m.viaStations = []*api.Location{nil}
	m.viaIndex = 0
	m.suggestions = []api.Location{{Name: "Olten"}, {Name: "Oltenweg"}}
	m.cursor = 0
	m.state = stateConnectionViaSuggestions

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm := newModel.(*Model)
	if nm.state != stateConnectionReady {
		t.Fatalf("expected %v got %v", stateConnectionReady, nm.state)
	}
	if nm.viaInputs[0].Value() != "Olten" {
		t.Errorf("via input not updated: %s", nm.viaInputs[0].Value())
	}
	if nm.cursor != 1 {
		t.Errorf("expected cursor on via, got %d", nm.cursor)
	}
}

func TestFromSuggestionSelectReturnsToReady(t *testing.T) {
	m := InitialModel()
	m.toStation = &api.Location{Name: "B"}
	m.suggestions = []api.Location{{Name: "A"}, {Name: "AA"}}
	m.cursor = 0
	m.state = stateConnectionFromSuggestions

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm := newModel.(*Model)
	if nm.state != stateConnectionReady {
		t.Fatalf("expected %v got %v", stateConnectionReady, nm.state)
	}
	if nm.cursor != 0 {
		t.Errorf("expected cursor on from, got %d", nm.cursor)
	}
	if nm.fromStation == nil || nm.fromStation.Name != "A" {
		t.Errorf("from station not set correctly")
	}
}
