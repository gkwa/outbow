package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/taylormonacelli/goldbug"
	"github.com/taylormonacelli/outbow"
	optmod "github.com/taylormonacelli/outbow/options"
)

func main() {
	options := optmod.Options{}

	flag.BoolVar(&options.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&options.Verbose, "v", false, "Enable verbose output (shorthand)")

	flag.StringVar(&options.LogFormat, "log-format", "", "Log format (text or json)")
	flag.StringVar(&options.StorageType, "storage", "db", "Storage type: db or json")
	flag.IntVar(&options.SubsetPercentage, "subset", 10, "Out of all pages, take a subset percentage of pages")

	flag.Parse()

	if options.Verbose || options.LogFormat != "" {
		if options.LogFormat == "json" {
			goldbug.SetDefaultLoggerJson(slog.LevelDebug)
		} else {
			goldbug.SetDefaultLoggerText(slog.LevelDebug)
		}
	}

	code := outbow.Main(options)
	os.Exit(code)
}
