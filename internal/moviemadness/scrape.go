package moviemadness

import (
	"context"
	"fmt"
	"strings"

	"github.com/drewfead/mmu/internal/core"
	"github.com/drewfead/mmu/internal/scraping"
	"github.com/gocolly/colly/v2"
)

type Scraper struct {
	BaseURL string
}

type SearchField int

const (
	Unknown SearchField = iota
	Title
	Genre
	NewReleases
	Director
	Actors
	Room
	Location
	Section
	All
	OutOfBounds
)

func (sf SearchField) String() string {
	switch sf {
	case Title:
		return "title"
	case Genre:
		return "genre"
	case NewReleases:
		return "newreleases"
	case Director:
		return "director"
	case Actors:
		return "actors"
	case Room:
		return "room"
	case Location:
		return "location"
	case Section:
		return "section"
	case All:
		return "all"
	default:
		return "unknown"
	}
}

const (
	searchResultClass         = ".filmInfo"
	searchResultTitleClass    = ".title"
	searchResultCategoryClass = ".category"
)

type ExtendedMovie struct {
	core.Movie
	Data map[string]any
}

func (sc *Scraper) Search(
	ctx context.Context,
	field SearchField,
	query string,
) ([]ExtendedMovie, error) {
	inQuery := strings.Builder{}
	queryFields := strings.Fields(query)
	for i, f := range strings.Fields(query) {
		inQuery.WriteString(f)
		if i != len(queryFields)-1 {
			inQuery.WriteString("+")
		}
	}

	searchURL := fmt.Sprintf("%s/search/?field=%s&query=%s", sc.BaseURL, field, inQuery.String())

	return scraping.SimpleGet(
		ctx,
		searchURL,
		map[string]string{"Referer": searchURL},
		searchResultClass,
		func(h *colly.HTMLElement) (ExtendedMovie, error) {
			title := h.ChildText(searchResultTitleClass + " > h3")
			categorySearchLinks := h.ChildAttrs(searchResultCategoryClass+" > a", "href")
			return ExtendedMovie{
				Movie: core.Movie{
					Title: title,
				},
				Data: map[string]any{
					"categorySearchLinks": categorySearchLinks,
				},
			}, nil
		},
	)
}
