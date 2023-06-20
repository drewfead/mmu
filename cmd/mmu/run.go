package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/drewfead/mmu/internal/commands"
)

func main() {
	app := &cli.App{
		Name:     "mmu",
		Usage:    "A utility for scraping websites for data about upcoming theatrical showings and home-video availability",
		Commands: commands.Scrapers,
	}
	if err := app.Run(os.Args); err != nil {
		zap.L().Fatal("Fatal error", zap.Error(err))
	}
}
