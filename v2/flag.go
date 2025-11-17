package main

import (
	"flag"
	"strings"

	"github.com/spf13/viper"
)

var versionFlagSet bool
var helpFlagSet bool

func initFlags(incomingFlags []string) {
	// Reset the sourcedir for when we're testing. Otherwise the global state
	// is preserved between the tests.
	AhoyConf.srcDir = ""

	// Reset flag detection
	versionFlagSet = false
	helpFlagSet = false

	// Normalize flags to single dash for parsing (convert --foo to -foo)
	// This is needed because Go's standard flag package doesn't support double dash
	normalizedFlags := make([]string, len(incomingFlags))
	for i, arg := range incomingFlags {
		if strings.HasPrefix(arg, "--") {
			normalizedFlags[i] = "-" + strings.TrimPrefix(arg, "--")
		} else {
			normalizedFlags[i] = arg
		}
	}

	// Parse the incoming flags using Go's standard flag package
	// This is needed for compatibility with the test suite and to
	// parse flags before cobra initialization
	tempFlags := flag.NewFlagSet("tempFlags", flag.ContinueOnError)
	tempFlags.StringVar(&sourcefile, "f", "", "specify the sourcefile")
	tempFlags.StringVar(&sourcefile, "file", "", "specify the sourcefile")
	tempFlags.BoolVar(&verbose, "v", false, "verbose output")
	tempFlags.BoolVar(&verbose, "verbose", false, "verbose output")

	// Add version and help flags that will be handled by cobra
	var versionFlag, helpFlag bool
	tempFlags.BoolVar(&versionFlag, "version", false, "print version")
	tempFlags.BoolVar(&helpFlag, "help", false, "print help")
	tempFlags.BoolVar(&helpFlag, "h", false, "print help")
	tempFlags.BoolVar(&bashCompletion, "generate-bash-completion", false, "")

	// Silently parse the flags - errors will be handled by cobra
	tempFlags.Parse(normalizedFlags)

	// Store whether version/help were requested
	versionFlagSet = versionFlag
	helpFlagSet = helpFlag

	// Update viper with parsed values
	if sourcefile != "" {
		viper.Set("file", sourcefile)
	}
	if verbose {
		viper.Set("verbose", verbose)
	}
}
