package hollywood_test

import (
	"context"
	"testing"

	"github.com/drewfead/mmu/internal/hollywood"
	"github.com/stretchr/testify/assert"
)

func Test_Real_Search(t *testing.T) {
	scraper := hollywood.Scraper{
		BaseURL: "https://hollywoodtheatre.org",
	}

	results, err := scraper.Calendar(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}
