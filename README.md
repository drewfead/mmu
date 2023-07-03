# mmu

A utility for scraping websites for data about upcoming theatrical showings and home-video availability

## Commands

### For Theatrical Releases

#### Get Hollywood Theatre Upcoming Showtimes

```
mmu hollywood-theatre
```

#### Get Hollywood Theatre Now Playing

```
mmu hollywood-theatre --now-playing
```

### For Home Video

#### Check Movie Madness Availability

```
mmu movie-madness [search terms ...]
```

### Global Flags

| Flag | Default | Usage |
|------|---------|-------|
| `--verbosity`, `-v` | info | Max log level to print when running, i.e. debug, info, warning, error |
| `--profile` | false | Output pprof profiling data to /tmp for the command run |
| `--output`, `-o` | json | Format in which to output the results of the command, i.e. json |
| `--help` | false | Print help text for the selected command |

## Development

### Running Locally

The CLI version of this module can be run from the project root using `go run` as follows:

```
go run ./cmd/mmu/run.go
```

This is the equivalent of running `mmu` with the installed binary, and supports all the same commands and flags.