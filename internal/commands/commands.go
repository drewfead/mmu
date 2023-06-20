package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"

	"github.com/drewfead/mmu/internal/core"
	"github.com/drewfead/mmu/internal/hollywood"
	"github.com/drewfead/mmu/internal/moviemadness"
)

var (
	profileFlag = &cli.BoolFlag{
		Name:  "profile",
		Usage: "Enable pprof profiling for this run",
		Value: false,
	}

	verbosityFlag = &cli.StringFlag{
		Name:  "verbosity",
		Usage: "Set the verbosity of the logger",
		Value: "info",
	}

	nowPlayingFlag = &cli.BoolFlag{
		Name:  "now-playing",
		Usage: "Scrape now playing movies",
		Value: false,
	}
)

func setup(ctx *cli.Context) []func() {
	zapCfg := zap.NewDevelopmentConfig()
	level, err := zap.ParseAtomicLevel(ctx.String(verbosityFlag.Name))
	if err != nil {
		log.Fatalf("failed to parse log level: %v", err)
	}
	zapCfg.Level = level
	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	zap.ReplaceGlobals(logger)
	maxprocs.Set(maxprocs.Logger(func(format string, args ...interface{}) {
		zap.L().Debug(fmt.Sprintf(format, args...))
	}))

	var out []func()

	if ctx.Bool(profileFlag.Name) {
		cpuProfile, err := os.Create("/tmp/cpu_profile.prof")
		if err != nil {
			log.Fatal(err)
		}

		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			log.Fatal(err)
		}

		out = append(out, func() {
			pprof.StopCPUProfile()
		})

		memProfile, err := os.Create("/tmp/memory_profile.prof")
		if err != nil {
			log.Fatal(err)
		}

		out = append(out, func() {
			memProfile.Close()
			runtime.GC()
			if err := pprof.WriteHeapProfile(memProfile); err != nil {
				zap.L().Error("Failed to write heap profile", zap.Error(err))
			}
		})
	}

	return out
}

func cleanup(ctx *cli.Context, steps ...func()) {
	for _, step := range steps {
		step()
	}
}

var Scrapers = []*cli.Command{
	{
		Name:     "hollywood-theatre",
		Usage:    "Find movies playing now or soon at the Hollywood Theatre",
		Category: "theatrical",
		Flags: []cli.Flag{
			verbosityFlag,
			profileFlag,
			nowPlayingFlag,
		},
		Action: func(c *cli.Context) error {
			cleanupSteps := setup(c)
			defer cleanup(c, cleanupSteps...)

			tz, err := time.LoadLocation("America/Los_Angeles")
			if err != nil {
				return err
			}

			s := &hollywood.Scraper{
				BaseURL:  "https://hollywoodtheatre.org",
				TimeZone: tz,
			}

			var movies []core.ExtendedMovie
			if c.Bool(nowPlayingFlag.Name) {
				movies, err = s.NowPlaying(context.Background())
			} else {
				movies, err = s.ComingSoon(context.Background())
			}
			if err != nil {
				return err
			}

			zap.L().Info("Found movies", zap.Any("movies", movies))
			return nil
		},
	},
	{
		Name:     "movie-madness",
		Usage:    "Find movies available to rent at Movie Madness",
		Category: "home-video",
		Flags: []cli.Flag{
			verbosityFlag,
			profileFlag,
		},
		ArgsUsage: "movie-madness [search term]",
		Action: func(c *cli.Context) error {
			cleanupSteps := setup(c)
			defer cleanup(c, cleanupSteps...)

			s := &moviemadness.Scraper{
				BaseURL: "https://moviemadness.org",
			}

			movies, err := s.Search(context.Background(), moviemadness.All, strings.Join(c.Args().Slice(), " "))
			if err != nil {
				return err
			}

			zap.L().Info("Found movies", zap.Any("movies", movies))
			return nil
		},
	},
}
