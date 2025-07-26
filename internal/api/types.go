package api

// JSON data models for the Swiss transport API

type Coordinate struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type Location struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Score      *float64   `json:"score"`
	Coordinate Coordinate `json:"coordinate"`
}

type LocationResponse struct {
	Stations []Location `json:"stations"`
}

type Prognosis struct {
	Platform string `json:"platform"`
}

type Stop struct {
	Station   Location  `json:"station"`
	Departure string    `json:"departure"`
	Arrival   string    `json:"arrival"`
	Delay     int       `json:"delay"`
	Platform  string    `json:"platform"`
	Prognosis Prognosis `json:"prognosis"`
}

type StationboardEntry struct {
	Stop     Stop   `json:"stop"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Number   string `json:"number"`
	To       string `json:"to"`
}

type StationboardResponse struct {
	Station      Location            `json:"station"`
	Stationboard []StationboardEntry `json:"stationboard"`
}

type Journey struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Number   string `json:"number"`
	Operator string `json:"operator"`
	To       string `json:"to"`
}

type Walk struct {
	Duration int `json:"duration"`
}

type Section struct {
	Journey   *Journey `json:"journey"`
	Walk      *Walk    `json:"walk"`
	Departure Stop     `json:"departure"`
	Arrival   Stop     `json:"arrival"`
}

type Connection struct {
	From struct {
		Station   Location `json:"station"`
		Departure string   `json:"departure"`
		Delay     int      `json:"delay"`
	} `json:"from"`
	To struct {
		Station Location `json:"station"`
		Arrival string   `json:"arrival"`
		Delay   int      `json:"delay"`
	} `json:"to"`
	Duration string    `json:"duration"`
	Sections []Section `json:"sections"`
}

type ConnectionsResponse struct {
	Connections []Connection `json:"connections"`
}
