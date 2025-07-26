package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"SBBuddy/internal/api"
	"SBBuddy/internal/ui"
)

func main() {
	args := insertDefaultRandom(os.Args[1:])

	randomVia := flag.Int("R", 0, "Get a random connection. Value specifies number of via stations")
	station := flag.String("T", "", "Lookup timetable for the given station")
	date := flag.String("d", "", "Date for lookup (YYYY-MM-DD or DD.MM.YYYY)")
	tm := flag.String("t", "", "Time for lookup (HH:mm)")
	arrival := flag.Bool("a", false, "Use arrival time instead of departure")

	var connections multiFlag
	flag.Var(&connections, "C", "Specify origin and destination; first and last are origin and destination, all others are via stations")

	flag.CommandLine.Parse(args)

	var randomFlag bool
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if f.Name == "R" {
			randomFlag = true
		}
	})
	if randomFlag {
		if *randomVia < 0 {
			fmt.Fprintln(os.Stderr, "Error: -R value must be >= 0")
			os.Exit(1)
		}

		client := api.NewClient(&http.Client{Timeout: 8 * time.Second})
		var from, to string
		var via []string

		if len(connections) == 0 {
			stations, err := api.RandomStations(*randomVia)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			from = stations[0]
			to = stations[len(stations)-1]
			if len(stations) > 2 {
				via = stations[1 : len(stations)-1]
			}
		} else {
			from = connections[0]
			exclude := []string{from}
			providedVia := []string{}
			if len(connections) >= 2 {
				to = connections[len(connections)-1]
				exclude = append(exclude, to)
				if len(connections) > 2 {
					providedVia = connections[1 : len(connections)-1]
					exclude = append(exclude, providedVia...)
				}
			}

			need := *randomVia
			if to == "" {
				need++
			}
			extras, err := api.RandomStationsExclude(need, exclude)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if to == "" {
				to = extras[len(extras)-1]
				extras = extras[:len(extras)-1]
			}
			via = append(providedVia, extras...)
		}

		sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		sp.Suffix = " Fetching..."
		sp.Start()
		cr, err := client.FetchConnections(context.Background(), from, to, via)
		sp.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(ui.FormatConnectionsTitle(from, via, to, ""))
		fmt.Print(ui.RenderConnectionsTable(cr))
		return
	}

	if len(connections) >= 2 {
		dateStr, err := ui.ParseDateInput(*date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date: %v\n", err)
			os.Exit(1)
		}
		timeStr, err := ui.ParseTimeInput(*tm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid time: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(&http.Client{Timeout: 8 * time.Second})
		sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		sp.Suffix = " Fetching..."
		sp.Start()

		var cr *api.ConnectionsResponse
		from := connections[0]
		to := connections[len(connections)-1]
		via := []string{}
		if len(connections) > 2 {
			via = connections[1 : len(connections)-1]
		}
		if *date != "" || *tm != "" {
			cr, err = client.FetchConnectionsAt(context.Background(), from, to, via, dateStr, timeStr, *arrival)
		} else {
			cr, err = client.FetchConnections(context.Background(), from, to, via)
		}
		sp.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		info := ""
		if *date != "" || *tm != "" {
			info = fmt.Sprintf("%s %s", ui.FormatDateDisplay(dateStr), timeStr)
		}
		fmt.Println(ui.FormatConnectionsTitle(from, via, to, info))
		fmt.Print(ui.RenderConnectionsTable(cr))
		return
	}

	if *station != "" {
		dateStr, err := ui.ParseDateInput(*date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date: %v\n", err)
			os.Exit(1)
		}
		timeStr, err := ui.ParseTimeInput(*tm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid time: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(&http.Client{Timeout: 8 * time.Second})
		sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		sp.Suffix = " Fetching..."
		sp.Start()

		var sb *api.StationboardResponse
		if *date != "" || *tm != "" {
			sb, err = client.FetchStationboardAt(context.Background(), *station, dateStr, timeStr)
		} else {
			sb, err = client.FetchStationboard(context.Background(), *station)
		}
		sp.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		title := fmt.Sprintf("Stationboard for %s", *station)
		if *date != "" || *tm != "" {
			title += fmt.Sprintf(" on %s %s", ui.FormatDateDisplay(dateStr), timeStr)
		}
		fmt.Println(title)
		fmt.Print(ui.RenderStationboardTable(sb))
		return
	}

	p := tea.NewProgram(ui.InitialModel())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func insertDefaultRandom(args []string) []string {
	for i := 0; i < len(args); i++ {
		if args[i] == "-R" {
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				args = append(args[:i+1], append([]string{"0"}, args[i+1:]...)...)
			}
			i++
		}
	}
	return args
}
