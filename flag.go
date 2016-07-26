package main

import (
	"flag"
	"github.com/codegangsta/cli"
	"os"
)

var globalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:        "verbose, v",
		Usage:       "Output extra details like the commands to be run.",
		EnvVar:      "AHOY_VERBOSE",
		Destination: &verbose,
	},
	cli.StringFlag{
		Name:        "file, f",
		Usage:       "Use a specific ahoy file.",
		Destination: &sourcefile,
	},
	cli.BoolFlag{
		Name:  "help, h",
		Usage: "show help",
	},
	cli.BoolFlag{
		Name:  "version",
		Usage: "print the version",
	},
	cli.BoolFlag{
		Name: "generate-bash-completion",
	},
}

// Sets flags for use by cli.flag using core flag module.
// I think we need this because we declare some flags on the fly at runtime?
func flagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

//TODO
func initFlags() {
	// Grab the global flags first ourselves so we can customize the yaml file loaded.
	tempFlags := flagSet("tempFlags", globalFlags)
	// Set the global flags using all the root Args beside the first one (which should be `ahoy`).
	// @frankcarey isn't 100% sure why this was needed, but it's related to being able to parse things at runtime.
	tempFlags.Parse(os.Args[1:])
}

// TODO
func overrideFlags(app *cli.App) {
	//Update the flags ourselves at runtime.
	app.Flags = globalFlags

	// TODO, not sure why these are set here, but they're related to hiding commands I think.
	app.HideVersion = true
	app.HideHelp = true
}
