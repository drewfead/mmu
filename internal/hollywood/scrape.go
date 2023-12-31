package hollywood

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/drewfead/mmu/internal/core"
	"github.com/drewfead/mmu/internal/scraping"
)

type Scraper struct {
	BaseURL  string
	TimeZone *time.Location

	session *scrapeSession
	sync.Mutex
}

func (s *Scraper) ComingSoon(ctx context.Context) ([]core.Movie, error) {
	ctx, span := otel.Tracer("hollywood.scraper").Start(ctx, "coming_soon")
	defer span.End()

	if err := s.initCurrent(ctx); err != nil {
		return nil, err
	}
	return s.calendar(ctx)
}

func (s *Scraper) NowPlaying(ctx context.Context) ([]core.Movie, error) {
	ctx, span := otel.Tracer("hollywood.scraper").Start(ctx, "now_playing")
	defer span.End()

	s.Lock()

	if s.session != nil && s.session.tab != "now_playing" {
		s.session = nil
		s.Unlock()
		return s.calendar(ctx)
	}

	s.Unlock()
	return s.calendar(ctx)
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
	ReflexID               string         `json:"reflexId"`
	ResolveLate            bool           `json:"resolveLate"`
	Selectors              []string       `json:"selectors"`
	TabID                  string         `json:"tabId"`
	XPathController        string         `json:"xpathController"`
	XPathElement           string         `json:"xpathElement"`
}

type cableReadyMessage struct {
	Identifier string          `json:"identifier"`
	Type       string          `json:"type"`
	CableReady bool            `json:"cableReady"`
	Operations json.RawMessage `json:"operations"`
}

type scrapeSession struct {
	sessionID string
	csrfToken string
	tab       string
}

const (
	kb = 1024
	mb = 1024 * kb
)

func (s *Scraper) sessionWSHandshake(
	ctx context.Context,
	c *colly.Collector,
	errs chan<- error,
	session chan<- *scrapeSession,
) func(r *colly.Response) {
	return func(r *colly.Response) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var csrfToken string
		for _, cookie := range c.Cookies(s.BaseURL) {
			if cookie.Name == "csrftoken" {
				csrfToken = cookie.Value
				break
			}
		}
		if csrfToken == "" {
			errs <- fmt.Errorf("failed to get csrfToken")
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
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			errs <- fmt.Errorf("failed to upgrade to WebSocket: %s", resp.Status)
			return
		}
		if err != nil {
			errs <- err
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "Done")

		if err := wsjson.Write(ctx, c, subscribeMessage{
			Type:        "subscribe",
			ChannelName: sockpuppetChannelName,
		}); err != nil {
			errs <- err
			return
		}

		var sessionID string

		var csm createSessionMessage
		if err := wsjson.Read(ctx, c, &csm); err != nil {
			errs <- err
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
			ReflexID:               uuid.New().String(),
			Selectors:              []string{"#eventGrid"},
			TabID:                  uuid.New().String(),
			XPathController:        "//*[@id='hwtController']/div[2]/div[2]",
			XPathElement:           "//*[@id='hwtController']/div[2]/div[2]",
		}); err != nil {
			errs <- err
			return
		}
		c.SetReadLimit(1 * mb)

		var crm cableReadyMessage
		if err := wsjson.Read(ctx, c, &crm); err != nil {
			errs <- err
			return
		}
		zap.L().Debug("cableReadyMessage", zap.Any("crm", crm))

		if sessionID == "" {
			errs <- fmt.Errorf("failed to get sessionID")
		}

		session <- &scrapeSession{
			sessionID: sessionID,
			csrfToken: csrfToken,
			tab:       "coming_soon",
		}
	}
}

func (s *Scraper) initCurrent(ctx context.Context) error {
	ctx, span := otel.Tracer("hollywood.scraper").Start(ctx, "init_current")
	defer span.End()

	s.Lock()
	defer s.Unlock()

	if s.session != nil && s.session.tab == "coming_soon" {
		return nil
	}

	errs := make(chan error)
	session := make(chan *scrapeSession, 1)

	c := colly.NewCollector(colly.Async(true))
	c.OnRequest(scraping.InjectRequestHeaders(map[string]string{
		"Referer": fmt.Sprintf("%s/coming-soon/", s.BaseURL),
	}))
	c.OnRequest(scraping.AddOutgoingContext(ctx))
	c.OnResponse(scraping.LogResponses(c))
	c.OnResponse(scraping.ReportBadResponses(s.BaseURL, nil))
	c.OnResponse(s.sessionWSHandshake(ctx, c, errs, session))

	go func() {
		if err := c.Visit(s.BaseURL); err != nil {
			errs <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			if err != nil {
				return err
			}
		case sess := <-session:
			if sess != nil {
				s.session = sess
				return nil
			}
		}
	}
}

func (s *Scraper) calendar(ctx context.Context) ([]core.Movie, error) {
	ctx, span := otel.Tracer("hollywood.scraper").Start(ctx, "calendar")
	defer span.End()

	s.Lock()
	defer s.Unlock()

	headers := map[string]string{
		"Referer": fmt.Sprintf("%s/coming-soon/", s.BaseURL),
	}

	if s.session != nil {
		headers["Cookie"] = fmt.Sprintf("csrftoken=%s; sessionid=%s", s.session.csrfToken, s.session.sessionID)
	}

	c := colly.NewCollector(colly.Async(true))
	hits := make(chan core.Movie)
	errs := make(chan error)
	c.OnRequest(scraping.InjectRequestHeaders(headers))
	c.OnResponse(scraping.LogResponses(c))
	c.OnResponse(scraping.ReportBadResponses(s.BaseURL, errs))
	c.OnHTML(".event-grid-item", s.parseMovie(hits))

	c.OnScraped(func(r *colly.Response) {
		close(hits)
		close(errs)
	})

	go func() {
		if err := c.Visit(s.BaseURL); err != nil {
			errs <- err
		}
	}()

	var out []core.Movie
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

func (s *Scraper) parseMovie(hits chan<- core.Movie) func(*colly.HTMLElement) {
	return func(e *colly.HTMLElement) {
		seriesName := strings.TrimSpace(e.ChildText(".event-grid-header > a.event_list__series_name"))
		seriesLink := e.ChildAttr(".event-grid-header > a.event_list__series_name", "href")
		if len(seriesLink) > 0 && !strings.HasPrefix(seriesLink, "https://") {
			seriesLink = fmt.Sprintf("%s/%s", s.BaseURL, strings.TrimPrefix(seriesLink, "/"))
		}
		dataEventID := e.ChildAttr(".event-grid-header > div > h3 > a", "data-event-id")
		title := strings.TrimSpace(e.ChildText(".event-grid-header > div > h3 > a"))
		if title == "" {
			// TODO: trap instances of unparseable movies
			return
		}
		images := make(map[string]string)
		e.ForEach(".event_list__image > a > picture > source", func(_ int, imgSrc *colly.HTMLElement) {
			images[imgSrc.Attr("type")] = imgSrc.Attr("srcset")
		})

		showtimes := s.parseShowtimes(e)
		if showtimes == nil {
			// TODO: trap instances of unparseable movies
			return
		}

		hollywoodScreening := core.Screening{
			Location:   "Hollywood Theatre",
			SeriesName: seriesName,
			SeriesLink: seriesLink,
			LinkURL:    fmt.Sprintf("%s/events/%s/", s.BaseURL, dataEventID),
		}

		if webpImg, hasWebP := images["image/webp"]; hasWebP {
			hollywoodScreening.ImageURLs = append(hollywoodScreening.ImageURLs, webpImg)
		} else if pngImg, hasPNG := images["image/png"]; hasPNG {
			hollywoodScreening.ImageURLs = append(hollywoodScreening.ImageURLs, pngImg)
		}

		for dayStr, times := range showtimes {
			for _, show := range times {
				var at time.Time
				if dt, ok := ParseDateTime(dayStr, show.time, s.TimeZone); ok {
					at = dt
				}
				// if datetime parsing fails, at will be time.Time{}, i.e. it.IsZero() will be true
				hollywoodScreening.Showtimes = append(hollywoodScreening.Showtimes, core.Showtime{
					At:      at,
					LinkURL: show.ticketLink,
				})
			}
		}

		out := core.Movie{
			Title:      title,
			Screenings: []core.Screening{hollywoodScreening},
		}

		hits <- out
	}
}

func (s *Scraper) parseShowtimes(e *colly.HTMLElement) map[string][]showtime {
	out := make(map[string][]showtime)
	e.ForEach(".event-grid-showtimes > div > div > div.carousel-item", func(_ int, carouselItem *colly.HTMLElement) {
		day := strings.TrimSpace(carouselItem.ChildText("h4.showtimes_date_header"))
		if day == "" {
			return
		}

		st := make([]showtime, 0)
		carouselItem.ForEach("div.showtime-square > a", func(_ int, showtimeSq *colly.HTMLElement) {
			ticketLink := showtimeSq.Attr("href")
			if len(ticketLink) > 0 && !strings.HasPrefix(ticketLink, "https://") {
				ticketLink = fmt.Sprintf("%s/%s", s.BaseURL, strings.TrimPrefix(ticketLink, "/"))
			}
			time := strings.TrimSpace(showtimeSq.Text)
			st = append(st, showtime{ticketLink: ticketLink, time: time})
		})

		if len(st) == 0 {
			return
		}

		if _, hasDay := out[day]; !hasDay {
			out[day] = st
		} else {
			out[day] = append(out[day], st...)
		}
	})

	if len(out) == 0 {
		return nil
	}
	return out
}
