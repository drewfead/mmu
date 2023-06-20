package core

import "time"

type Showtime struct {
	At      time.Time
	LinkURL string
}

type Screening struct {
	Location   string
	SeriesName string
	SeriesLink string
	LinkURL    string
	Showtimes  []Showtime
	ImageURLs  []string
}

type FormatSpecificAvailability struct {
	Format string
	Count  int
}

type Availability struct {
	Location string
	Formats  []FormatSpecificAvailability
}

type Movie struct {
	Title      string
	Screenings []Screening
	// Availability []Availability
}

type ExtendedMovie struct {
	Movie
	Data map[string]any
}
