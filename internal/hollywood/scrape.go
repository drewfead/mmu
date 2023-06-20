package hollywood

import (
	"context"
	"fmt"

	"github.com/drewfead/mmu/internal/core"
	"github.com/drewfead/mmu/internal/scraping"
	"github.com/gocolly/colly/v2"
)

type ExtendedMovie struct {
	core.Movie
	Data map[string]any
}

type Scraper struct {
	BaseURL string
}

type showtime struct {
	ticketLink string
	time       string
}

func (s *Scraper) Calendar(ctx context.Context) ([]ExtendedMovie, error) {
	c := colly.NewCollector(colly.Async(true))
	hits := make(chan ExtendedMovie)
	errs := make(chan error)
	c.OnResponse(scraping.LogResponses())
	c.OnResponse(scraping.ReportBadResponses(s.BaseURL, errs))
	c.OnRequest(scraping.InjectRequestHeaders(map[string]string{
		"Referer": fmt.Sprintf("%s/coming-soon/", s.BaseURL),
	}))

	c.OnHTML(".event-grid-item", func(e *colly.HTMLElement) {
		seriesName := e.ChildText(".event-grid-header > a.event_list__series_name")
		seriesLink := e.ChildAttr(".event-grid-header > a.event_list__series_name", "href")
		dataEventID := e.ChildAttr(".event-grid-header > div > h3 > a", "data-event-id")
		title := e.ChildText(".event-grid-header > div > h3 > a")
		images := make(map[string]string)
		e.ForEach(".event_list__image > a > picture > source", func(_ int, imgSrc *colly.HTMLElement) {
			images[imgSrc.Attr("type")] = imgSrc.Attr("srcset")
		})
		showtimes := make(map[string][]showtime)
		e.ForEach(".event-grid-showtimes > div > div > div.carousel-item", func(_ int, carouselItem *colly.HTMLElement) {
			day := carouselItem.ChildText("h4.showtimes_date_header")
			if day == "" {
				return
			}
			if _, hasDay := showtimes[day]; !hasDay {
				showtimes[day] = make([]showtime, 0)
			}
			carouselItem.ForEach("div.showtime-square > a", func(_ int, showtimeSq *colly.HTMLElement) {
				ticketLink := showtimeSq.Attr("href")
				time := showtimeSq.Text
				showtimes[day] = append(showtimes[day], showtime{ticketLink: ticketLink, time: time})
			})
		})
		hits <- ExtendedMovie{
			Movie: core.Movie{
				Title: title,
			},
			Data: map[string]any{
				"seriesName":  seriesName,
				"seriesLink":  seriesLink,
				"dataEventID": dataEventID,
				"images":      images,
				"showtimes":   showtimes,
			},
		}
	})

	c.OnScraped(func(r *colly.Response) {
		close(hits)
		close(errs)
	})

	go func() {
		c.Visit(s.BaseURL)
	}()

	var out []ExtendedMovie
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errs:
			if err != nil {
				return nil, err
			}
		case hit, more := <-hits:
			if !more {
				return out, nil
			}
			out = append(out, hit)
		}
	}
}
