package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	_ "go.uber.org/automaxprocs"

	"github.com/drewfead/mmu/internal/commands"
)

func main() {
	app := &cli.App{
		Name:     "mmu",
		Usage:    "A utility for scraping websites for data about upcoming theatrical showings and home-video availability",
		Commands: commands.Scrapers,
	}
	if err := app.Run(os.Args); err != nil {
		zap.L().Fatal("Failed to start perf test")
	}
}
