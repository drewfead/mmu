package moviemadness_test

import (
	"context"
	"testing"

	"github.com/drewfead/mmu/internal/moviemadness"
	"github.com/stretchr/testify/assert"
)

func Test_Real_Search(t *testing.T) {
	scraper := moviemadness.Scraper{
		BaseURL: "https://moviemadness.org",
	}

	results, err := scraper.Search(context.Background(), moviemadness.Title, "The Matrix")
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}
