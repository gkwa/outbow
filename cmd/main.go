package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/taylormonacelli/goldbug"
	"github.com/taylormonacelli/outbow"
)

// Options struct to hold command line options
type Options struct {
	Verbose     bool
	LogFormat   string
	StorageType string
}

func main() {
	// Create an instance of the Options struct
	options := Options{}

	// Define command line flags and bind them to the corresponding fields in the Options struct
	flag.BoolVar(&options.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&options.Verbose, "v", false, "Enable verbose output (shorthand)")

	flag.StringVar(&options.LogFormat, "log-format", "", "Log format (text or json)")
	flag.StringVar(&options.StorageType, "storage", "db", "Storage type: db or json")

	flag.Parse()

	if options.Verbose || options.LogFormat != "" {
		if options.LogFormat == "json" {
			goldbug.SetDefaultLoggerJson(slog.LevelDebug)
		} else {
			goldbug.SetDefaultLoggerText(slog.LevelDebug)
		}
	}

	// Pass the options struct to the Main function
	code := outbow.Main(options.StorageType)
	os.Exit(code)
}
