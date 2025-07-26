package ui

import "testing"

func TestActiveKeys(t *testing.T) {
	tests := []struct {
		state           appState
		backEnabled     bool
		qrEnabled       bool
		refreshEnabled  bool
		dateTimeEnabled bool
		addViaEnabled   bool
		modifyEnabled   bool
	}{
		{stateMenu, false, false, false, false, false, false},
		{stateStationInput, true, false, false, false, false, false},
		{stateShowStationboard, true, false, true, true, false, false},
		{stateShowConnections, true, false, true, true, false, true},
		{stateShowConnectionDetails, true, true, true, false, false, false},
		{stateConnectionInputFrom, true, false, false, false, true, false},
		{stateConnectionInputTo, true, false, false, false, false, false},
		{stateConnectionInputVia, true, false, false, false, true, false},
		{stateConnectionReady, true, false, false, false, true, false},
	}
	m := InitialModel()
	for _, tt := range tests {
		m.state = tt.state
		k := m.activeKeys()
		if k.Back.Enabled() != tt.backEnabled {
			t.Errorf("state %v: back enabled=%v, want %v", tt.state, k.Back.Enabled(), tt.backEnabled)
		}
		if k.QR.Enabled() != tt.qrEnabled {
			t.Errorf("state %v: qr enabled=%v, want %v", tt.state, k.QR.Enabled(), tt.qrEnabled)
		}
		if k.Refresh.Enabled() != tt.refreshEnabled {
			t.Errorf("state %v: refresh enabled=%v, want %v", tt.state, k.Refresh.Enabled(), tt.refreshEnabled)
		}
		if k.DateTime.Enabled() != tt.dateTimeEnabled {
			t.Errorf("state %v: datetime enabled=%v, want %v", tt.state, k.DateTime.Enabled(), tt.dateTimeEnabled)
		}
		if k.AddVia.Enabled() != tt.addViaEnabled {
			t.Errorf("state %v: addVia enabled=%v, want %v", tt.state, k.AddVia.Enabled(), tt.addViaEnabled)
		}
		if k.Modify.Enabled() != tt.modifyEnabled {
			t.Errorf("state %v: modify enabled=%v, want %v", tt.state, k.Modify.Enabled(), tt.modifyEnabled)
		}
	}
}
