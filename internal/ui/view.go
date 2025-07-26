package ui

import (
	"fmt"
	"strings"

	api "SBBuddy/internal/api"
)

func (m *Model) View() string {
	helpView := "\n\n" + m.help.View(m.activeKeys())
	switch m.state {
	case stateMenu:
		s := "🚂 Swiss Transport Timetable\n\n"
		s += "Select an option:\n\n"
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
		return s + helpView

	case stateLoadingConnections:
		return fmt.Sprintf("\n%s Loading...", m.spinner.View()) + helpView

	case stateStationInput:
		return "Enter station for timetable:\n\n" + m.stationInput.View() + helpView

	case stateStationSuggestions:
		s := "Multiple stations found. Select one:\n\n"
		for i, station := range m.suggestions {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, station.Name)
		}
		return s + helpView

	case stateConnectionInputFrom:
		fromText := ""
		if m.fromStation != nil {
			fromText = fmt.Sprintf(" (selected: %s)", m.fromStation.Name)
		}
		return fmt.Sprintf("Enter departure station%s:\n\n%s", fromText, m.fromInput.View()) + helpView

	case stateConnectionFromSuggestions:
		s := fmt.Sprintf("Multiple departure stations found. Select one:\n\n")
		for i, station := range m.suggestions {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, station.Name)
		}
		return s + helpView

	case stateConnectionInputTo:
		return renderConnectionInputs(m, false) + helpView

	case stateConnectionToSuggestions:
		s := "Multiple arrival stations found. Select one:\n\n"
		for i, station := range m.suggestions {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, station.Name)
		}
		return renderConnectionInputs(m, false) + "\n\n" + s + helpView

	case stateConnectionInputVia:
		return renderConnectionInputs(m, true) + helpView

	case stateConnectionViaSuggestions:
		s := "Multiple via stations found. Select one:\n\n"
		for i, station := range m.suggestions {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, station.Name)
		}
		return renderConnectionInputs(m, true) + "\n\n" + s + helpView

	case stateConnectionReady:
		return renderConnectionSummary(m) + helpView

	case stateDateTimeInput:
		s := "Select your date and time:\n\n" + m.dateTimeView()
		return s + helpView

	case stateShowStationboard:
		if m.err != nil {
			return fmt.Sprintf("❌ %v", m.err) + helpView
		}
		if m.stationboard == nil || len(m.stationboard.Stationboard) == 0 {
			stationName := "Unknown"
			if m.selectedStation != nil {
				stationName = m.selectedStation.Name
			}
			return fmt.Sprintf("📋 No departures found for %s", stationName) + helpView
		}
		table := renderStationboardTable(m.stationboard)
		info := ""
		if !m.lastSearchTime.IsZero() {
			mode := "Depart"
			if m.lastSearchArrival {
				mode = "Arrive"
			}
			info = fmt.Sprintf(" (%s %s %s)", mode, m.lastSearchTime.Format("Mon 02.01.2006"), m.lastSearchTime.Format("15:04"))
		}
		return fmt.Sprintf("📋 Stationboard for %s%s:\n\n%s", m.stationboard.Station.Name, info, table) + helpView

	case stateShowConnections:
		if m.err != nil {
			return fmt.Sprintf("❌ %v", m.err) + helpView
		}
		if m.connections == nil || len(m.connections.Connections) == 0 {
			fromName := "Unknown"
			toName := "Unknown"
			if m.fromStation != nil {
				fromName = m.fromStation.Name
			}
			if m.toStation != nil {
				toName = m.toStation.Name
			}
			return fmt.Sprintf("🔍 No connections found from %s to %s", fromName, toName) + helpView
		}

		fromName := "Unknown"
		toName := "Unknown"
		if len(m.connections.Connections) > 0 {
			fromName = m.connections.Connections[0].From.Station.Name
			toName = m.connections.Connections[0].To.Station.Name
		}

		// Enhanced connections list with selection cursor
		info := ""
		if !m.lastSearchTime.IsZero() {
			mode := "Depart"
			if m.lastSearchArrival {
				mode = "Arrive"
			}
			info = fmt.Sprintf("%s %s %s", mode, m.lastSearchTime.Format("Mon 02.01.2006"), m.lastSearchTime.Format("15:04"))
		}
		s := renderConnectionsHeader(fromName, m.viaNames(), toName, info) + "\n\n"
		s += m.connTable.View()
		return s + helpView

	case stateShowConnectionDetails:
		if m.selectedConnection == nil {
			return "No connection selected" + helpView
		}
		return renderConnectionDetails(m.selectedConnection) + helpView

	case stateShowConnectionQR:
		return fmt.Sprintf(
			"🔗 Scan to open in SBB timetable:\n\n%s", m.qrCode,
		) + helpView

	default:
		return ""
	}
}

func renderConnectionInputs(m *Model, editingVia bool) string {
	var sb strings.Builder
	if m.fromStation != nil {
		sb.WriteString(fmt.Sprintf("From: %s\n", m.fromStation.Name))
	} else {
		sb.WriteString("From: ?\n")
	}
	for i, inp := range m.viaInputs {
		label := fmt.Sprintf("Via %d: ", i+1)
		if editingVia && m.viaIndex == i {
			sb.WriteString(label + inp.View() + "\n")
			continue
		}
		if i < len(m.viaStations) && m.viaStations[i] != nil {
			sb.WriteString(label + m.viaStations[i].Name + "\n")
		} else {
			sb.WriteString(label + inp.View() + "\n")
		}
	}
	if m.state == stateConnectionInputTo {
		sb.WriteString("To: " + m.toInput.View())
	} else if m.toStation != nil {
		sb.WriteString("To: " + m.toStation.Name)
	} else {
		sb.WriteString("To: " + m.toInput.View())
	}
	return sb.String()
}

func renderConnectionSummary(m *Model) string {
	var lines []string
	from := "?"
	if m.fromStation != nil {
		from = m.fromStation.Name
	}
	lines = append(lines, fmt.Sprintf("From: %s", from))
	for i := range m.viaInputs {
		name := "?"
		if i < len(m.viaStations) && m.viaStations[i] != nil {
			name = m.viaStations[i].Name
		}
		lines = append(lines, fmt.Sprintf("Via %d: %s", i+1, name))
	}
	to := "?"
	if m.toStation != nil {
		to = m.toStation.Name
	}
	lines = append(lines, fmt.Sprintf("To: %s", to))
	lines = append(lines, "Search")

	var sb strings.Builder
	for i, line := range lines {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, line))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func renderConnectionDetails(conn *api.Connection) string {
	if conn == nil {
		return ""
	}

	var s strings.Builder

	// Header with journey overview
	s.WriteString("🚂 Connection Details 🚂\n")
	s.WriteString("═══════════════════════════════════════════════════════════\n\n")

	// Journey summary
	totalDuration := conn.Duration
	if totalDuration == "" {
		totalDuration = durationBetween(conn.From.Departure, conn.To.Arrival)
	} else {
		totalDuration = parseDuration(totalDuration)
	}

	changes := countChanges(conn.Sections)

	s.WriteString(fmt.Sprintf("📍 %s → %s\n", conn.From.Station.Name, conn.To.Station.Name))
	var delays []string
	if conn.From.Delay > 0 {
		delays = append(delays, formatDelay(conn.From.Delay))
	}
	if conn.To.Delay > 0 {
		delays = append(delays, formatDelay(conn.To.Delay))
	}
	var delayStr string
	if len(delays) > 0 {
		delayStr = fmt.Sprintf(", Delay: %s", strings.Join(delays, "/"))
	}
	s.WriteString(fmt.Sprintf("🕐 %s → %s (Duration: %s%s)\n",
		formatISOTime(conn.From.Departure),
		formatISOTime(conn.To.Arrival),
		totalDuration,
		delayStr))
	s.WriteString(fmt.Sprintf("🔄 %s\n\n", formatChanges(changes)))

	// Detailed journey sections
	s.WriteString("Journey Details:\n")
	s.WriteString("───────────────────────────────────────────────────────────\n")

	for i, section := range conn.Sections {
		// Walking section
		if section.Walk != nil {
			mins := section.Walk.Duration / 60
			walkDuration := fmt.Sprintf("%d min", mins)
			s.WriteString(fmt.Sprintf("🚶 Walk from %s to %s (%s)\n\n",
				section.Departure.Station.Name,
				section.Arrival.Station.Name,
				walkDuration))
			continue
		}

		// Transport section
		if section.Journey != nil {
			// Departure info
			depPlatform := section.Departure.Platform
			if depPlatform == "" {
				depPlatform = section.Departure.Prognosis.Platform
			}

			// Arrival info
			arrPlatform := section.Arrival.Platform
			if arrPlatform == "" {
				arrPlatform = section.Arrival.Prognosis.Platform
			}

			// Train info
			trainLabel := fmt.Sprintf("%s%s", section.Journey.Category, section.Journey.Number)

			s.WriteString(fmt.Sprintf("🚊 %s towards %s\n", trainLabel, section.Journey.To))
			delayDepStr := ""
			if section.Departure.Delay > 0 {
				delayDepStr = fmt.Sprintf(", Delay %s", formatDelay(section.Departure.Delay))
			}
			delayArrStr := ""
			if section.Arrival.Delay > 0 {
				delayArrStr = fmt.Sprintf(", Delay %s", formatDelay(section.Arrival.Delay))
			}
			depLabel := ""
			if depPlatform != "" {
				depLabel = fmt.Sprintf("Platform %s", depPlatform)
			}
			arrLabel := ""
			if arrPlatform != "" {
				arrLabel = fmt.Sprintf("Platform %s", arrPlatform)
			}
			s.WriteString(fmt.Sprintf("   Depart: %s from %s (%s%s)\n",
				formatISOTime(section.Departure.Departure),
				section.Departure.Station.Name,
				depLabel,
				delayDepStr))
			s.WriteString(fmt.Sprintf("   Arrive: %s at %s (%s%s)\n",
				formatISOTime(section.Arrival.Arrival),
				section.Arrival.Station.Name,
				arrLabel,
				delayArrStr))

			// Calculate section duration
			sectionDuration := durationBetween(section.Departure.Departure, section.Arrival.Arrival)
			if sectionDuration != "-" {
				s.WriteString(fmt.Sprintf("   Duration: %s\n", sectionDuration))
			}

			if i < len(conn.Sections)-1 {
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n───────────────────────────────────────────────────────────\n")

	return s.String()
}

func renderConnectionsHeader(from string, via []string, to, info string) string {
	var lines []string
	lines = append(lines, "🔍 Connections")
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
