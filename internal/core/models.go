package core

import "time"

type Showtime struct {
	At      time.Time `json:"at"`
	LinkURL string    `json:"link_url,omitempty"`
}

type Screening struct {
	Location   string     `json:"location,omitempty"`
	SeriesName string     `json:"series_name,omitempty"`
	SeriesLink string     `json:"series_link,omitempty"`
	LinkURL    string     `json:"link_url,omitempty"`
	Showtimes  []Showtime `json:"showtimes,omitempty"`
	ImageURLs  []string   `json:"image_urls,omitempty"`
}

type FormatSpecificAvailability struct {
	Format string `json:"format"`
	Count  int    `json:"count"`
}

type Availability struct {
	Location string                       `json:"location"`
	Formats  []FormatSpecificAvailability `json:"formats"`
}

type Movie struct {
	Title      string      `json:"title"`
	Screenings []Screening `json:"screenings"`
	// Availability []Availability
	Data map[string]any `json:"data,omitempty"`
}
