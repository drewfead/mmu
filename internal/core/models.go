package core

import "time"

type Showtime struct {
	At time.Time
}

type Screening struct {
	Location  string
	Showtimes []Showtime
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
	Title        string
	Screenings   []Screening
	Availability []Availability
}
