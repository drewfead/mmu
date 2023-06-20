package scraping

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gocolly/colly/v2"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func SimpleGet[OUT any](
	ctx context.Context,
	url string,
	requestHeaders map[string]string,
	elementSelector string,
	transformElement func(*colly.HTMLElement) (OUT, error),
) ([]OUT, error) {
	ctx, span := otel.Tracer("scraping").Start(ctx, "simple_get")
	defer span.End()

	c := colly.NewCollector(colly.Async(true))
	hits := make(chan OUT)
	errs := make(chan error)
	c.OnRequest(InjectRequestHeaders(requestHeaders))
	c.OnRequest(AddOutgoingContext(ctx))
	c.OnResponse(LogResponses(c))
	c.OnResponse(ReportBadResponses(url, errs))
	c.OnHTML(elementSelector, func(e *colly.HTMLElement) {
		hit, err := transformElement(e)
		if err != nil {
			errs <- err
		} else {
			hits <- hit
		}
	})
	c.OnScraped(func(r *colly.Response) {
		close(hits)
		close(errs)
	})

	go func() {
		c.Visit(url)
	}()

	var out []OUT
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

func ReportBadResponses(url string, errorChannel chan<- error) func(r *colly.Response) {
	return func(r *colly.Response) {
		if r.StatusCode != http.StatusOK {
			errorChannel <- fmt.Errorf("got status code %d from %s", r.StatusCode, url)
		}
	}
}

func LogResponses(c *colly.Collector) func(r *colly.Response) {
	return func(r *colly.Response) {
		cookies := c.Cookies(r.Request.URL.String())
		zap.L().Debug("response", zap.Int("status", r.StatusCode), zap.String("body", string(r.Body)), zap.Any("cookies", cookies))
	}
}

func InjectRequestHeaders(headers map[string]string) func(r *colly.Request) {
	return func(r *colly.Request) {
		for k, v := range headers {
			r.Headers.Set(k, v)
		}
	}
}

func AddOutgoingContext(ctx context.Context) func(r *colly.Request) {
	return func(r *colly.Request) {
		go func() {
			<-ctx.Done()
			r.Abort()
		}()
	}
}
