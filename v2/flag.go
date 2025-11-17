package main

import (
	"flag"

	"github.com/spf13/viper"
)

func initFlags(incomingFlags []string) {
	// Reset the sourcedir for when we're testing. Otherwise the global state
	// is preserved between the tests.
	AhoyConf.srcDir = ""

	// Parse the incoming flags using Go's standard flag package
	// This is needed for compatibility with the test suite and to
	// parse flags before cobra initialization
	tempFlags := flag.NewFlagSet("tempFlags", flag.ContinueOnError)
	tempFlags.StringVar(&sourcefile, "f", "", "specify the sourcefile")
	tempFlags.StringVar(&sourcefile, "file", "", "specify the sourcefile")
	tempFlags.BoolVar(&verbose, "v", false, "verbose output")
	tempFlags.BoolVar(&verbose, "verbose", false, "verbose output")
	tempFlags.BoolVar(&bashCompletion, "generate-bash-completion", false, "")

	// Silently parse the flags - errors will be handled by cobra
	tempFlags.Parse(incomingFlags)

	// Update viper with parsed values
	if sourcefile != "" {
		viper.Set("file", sourcefile)
	}
	if verbose {
		viper.Set("verbose", verbose)
	}
}
