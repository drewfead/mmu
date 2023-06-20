package hollywood_test

import (
	"context"
	"testing"
	"time"

	"github.com/drewfead/mmu/internal/hollywood"
	"github.com/stretchr/testify/assert"
)

func Test_Real_ComingSoon(t *testing.T) {
	tz, err := time.LoadLocation("America/Los_Angeles")
	assert.NoError(t, err)
	scraper := hollywood.Scraper{
		BaseURL:  "https://hollywoodtheatre.org",
		TimeZone: tz,
	}

	start := time.Now()
	results, err := scraper.ComingSoon(context.Background())
	took := time.Since(start).Milliseconds()

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Positive(t, took)
}

func Test_Real_NowPlaying(t *testing.T) {
	tz, err := time.LoadLocation("America/Los_Angeles")
	assert.NoError(t, err)
	scraper := hollywood.Scraper{
		BaseURL:  "https://hollywoodtheatre.org",
		TimeZone: tz,
	}

	start := time.Now()
	results, err := scraper.NowPlaying(context.Background())
	took := time.Since(start).Milliseconds()

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Positive(t, took)
}
