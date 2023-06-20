package hollywood

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/drewfead/mmu/internal/core"
	"github.com/drewfead/mmu/internal/scraping"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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

const (
	sockpuppetChannelName = "{\"channel\":\"StimulusReflex::Channel\"}"
)

type subscribeMessage struct {
	Type        string `json:"type"`
	ChannelName string `json:"channelName"`
}

type createSessionMessage struct {
	Type     string `json:"type"`
	MetaType string `json:"meta_type"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type changeTargetMessage struct {
	Target                 string         `json:"target"`
	Args                   []string       `json:"args"`
	URL                    string         `json:"url"`
	Identifier             string         `json:"identifier"`
	Attrs                  map[string]any `json:"attrs"`
	Dataset                map[string]any `json:"dataset"`
	FormData               string         `json:"formData"`
	PermanentAttributeName string         `json:"permanentAttributeName"`
	ReflexController       string         `json:"reflexController"`
	ReflexId               string         `json:"reflexId"`
	ResolveLate            bool           `json:"resolveLate"`
	Selectors              []string       `json:"selectors"`
	TabID                  string         `json:"tabId"`
	XPathController        string         `json:"xpathController"`
	XPathElement           string         `json:"xpathElement"`
}

type cableReadyMessage struct {
	Identifier string         `json:"identifier"`
	Type       string         `json:"type"`
	CableReady bool           `json:"cableReady"`
	Operations map[string]any `json:"operations"`
}

func (s *Scraper) Calendar(ctx context.Context) ([]ExtendedMovie, error) {
	initSession := colly.NewCollector()
	initSession.OnRequest(scraping.InjectRequestHeaders(map[string]string{
		"Referer": fmt.Sprintf("%s/coming-soon/", s.BaseURL),
	}))
	initSession.OnResponse(scraping.LogResponses(initSession))
	initSession.OnResponse(scraping.ReportBadResponses(s.BaseURL, nil))
	var failedInit error
	var csrfToken string
	var sessionID string
	initSession.OnResponse(func(r *colly.Response) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for _, cookie := range initSession.Cookies(s.BaseURL) {
			if cookie.Name == "csrftoken" {
				csrfToken = cookie.Value
				break
			}
		}
		if csrfToken == "" {
			failedInit = fmt.Errorf("failed to get csrfToken")
			return
		}
		schemeless := strings.TrimPrefix(s.BaseURL, "https://")
		c, resp, err := websocket.Dial(ctx, fmt.Sprintf("wss://%s/ws/sockpuppet-sync", schemeless), &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Cookie": []string{fmt.Sprintf("csrftoken=%s", csrfToken)},
				// required to avoid 403
				"Origin": []string{s.BaseURL},
			},
			CompressionMode: websocket.CompressionContextTakeover,
		})
		if resp.StatusCode != http.StatusSwitchingProtocols {
			failedInit = fmt.Errorf("failed to upgrade to WebSocket: %s", resp.Status)
			return
		}
		if err != nil {
			failedInit = err
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "Done")

		if err := wsjson.Write(ctx, c, subscribeMessage{
			Type:        "subscribe",
			ChannelName: sockpuppetChannelName,
		}); err != nil {
			failedInit = err
			return
		}

		var csm createSessionMessage
		if err := wsjson.Read(ctx, c, &csm); err != nil {
			failedInit = err
			return
		}
		if csm.Key == "sessionid" {
			sessionID = csm.Value
		}

		if err := wsjson.Write(ctx, c, changeTargetMessage{
			Target:     "EventsReflex#toggle_coming_soon",
			Args:       []string{},
			URL:        fmt.Sprintf("%s/", s.BaseURL),
			Identifier: sockpuppetChannelName,
			Attrs: map[string]any{
				"checked":          false,
				"data-controller":  "events",
				"data-reflex-root": "#eventGrid",
				"selected":         false,
				"tag_name":         "DIV",
			},
			Dataset: map[string]any{
				"dataset": map[string]any{
					"data-controller":  "events",
					"data-reflex-root": "#eventGrid",
				},
				"datasetAll": struct{}{},
			},
			PermanentAttributeName: "data-reflex-permanent",
			ReflexController:       "events",
			ReflexId:               uuid.New().String(),
			Selectors:              []string{"#eventGrid"},
			TabID:                  uuid.New().String(),
			XPathController:        "//*[@id='hwtController']/div[2]/div[2]",
			XPathElement:           "//*[@id='hwtController']/div[2]/div[2]",
		}); err != nil {
			failedInit = err
			return
		}

		c.SetReadLimit(1024 * 1024 * 1024)

		var crm cableReadyMessage
		if err := wsjson.Read(ctx, c, &crm); err != nil {
			failedInit = err
			return
		}
		zap.L().Debug("cableReadyMessage", zap.Any("crm", crm))
	})

	initSession.Visit(s.BaseURL)

	if failedInit != nil {
		return nil, failedInit
	}

	if sessionID == "" {
		return nil, fmt.Errorf("failed to get sessionID")
	}

	c := colly.NewCollector(colly.Async(true))
	hits := make(chan ExtendedMovie)
	errs := make(chan error)
	c.OnRequest(scraping.InjectRequestHeaders(map[string]string{
		"Referer": fmt.Sprintf("%s/coming-soon/", s.BaseURL),
		"Cookie":  fmt.Sprintf("csrftoken=%s; sessionid=%s", csrfToken, sessionID),
	}))
	c.OnResponse(scraping.LogResponses(c))
	c.OnResponse(scraping.ReportBadResponses(s.BaseURL, errs))

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
