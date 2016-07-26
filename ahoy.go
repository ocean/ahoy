package main

import (
	"flag"
	"fmt"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//Creates a structure for Config type
type Config struct {
	Usage    string
	AhoyAPI  string
	Version  string
	Commands map[string]Command
}

//Creates a structure for Config's commands
type Command struct {
	Description string
	Usage       string
	Cmd         string
	Hide        bool
	Imports     []string
}

var app *cli.App
var sourcedir string
var sourcefile string
var args []string
var verbose bool
var bashCompletion bool

// Prints the error and exits if it's a fatal error
func logger(errType string, text string) {
	err_text := ""
	if (errType == "error") || (errType == "fatal") || (verbose == true) {
		err_text = "AHOY! [" + errType + "] ==> " + text + "\n"
		log.Print(err_text)
	}
	if errType == "fatal" {
		panic(err_text)
	}
}

//Gets the path of the .ahoy.yml file if it exist
func getConfigPath(sourcefile string) (string, error) {
	var err error

	// If a specific source file was set, then try to load it directly.
	if sourcefile != "" {
		//Stat method returns the FileInfo structure describing the file. It returns an error if file doesn't exist.
		if _, err := os.Stat(sourcefile); err == nil {
			return sourcefile, err
		} else {
			logger("fatal", "An ahoy config file was specified using -f to be at "+sourcefile+" but couldn't be found. Check your path.")
		}
	}

	//Getwd method returns the present working directory.
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for dir != "/" && err == nil {
		ymlpath := filepath.Join(dir, ".ahoy.yml")
		//log.Println(ymlpath)
		if _, err := os.Stat(ymlpath); err == nil {
			//log.Println("found: ", ymlpath )
			return ymlpath, err
		}
		// Chop off the last part of the path.
		dir = path.Dir(dir)
	}
	return "", err
}

func getConfig(sourcefile string) (Config, error) {

	yamlFile, err := ioutil.ReadFile(sourcefile)
	if err != nil {
		logger("fatal", "An ahoy config file couldn't be found in your path. You can create an example one by using 'ahoy init'.")
	}

	var config Config
	// Extract the yaml file into the config varaible.
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	// All ahoy files (and imports) must specify the ahoy version.
	// This is so we can support backwards compatability in the future.
	if config.AhoyAPI != "v2" {
		logger("fatal", "Ahoy only supports API version 'v2', but '"+config.AhoyAPI+"' given in "+sourcefile)
	}

	return config, err
}
// TODO
func getSubCommands(includes []string) []cli.Command {
	subCommands := []cli.Command{}
	//return empty subcommands if no includes are passed.
	if 0 == len(includes) {
		return subCommands
	}
	commands := map[string]cli.Command{}

	// Iterate over each include.
	for _, include := range includes {
		// Should this exist? Why not fail when a line is empty?
		// Probably left over code from when this was parsing strings.
		if len(include) == 0 {
			continue
		}
		// Prepend the current ahoy.yml file path
		// for relative paths.
		if include[0] != "/"[0] || include[0] != "~"[0] {
			include = filepath.Join(sourcedir, include)
		}

		// Check if the full include path exists or not.
		if _, err := os.Stat(include); err != nil {
			//Skipping files that cannot be loaded allows us to separate
			//subcommands into public and private.

			// TODO: Fail unless a path starts with a "?",
			// which we use to signal that a missing file should
			// not cause an error, and should be skipped instead.
			continue
		}
		// Get the sub commands from the include file.
		// Reuses the base getConfig().
		config, _ := getConfig(include)
		includeCommands := getCommands(config)
		// Override existing commands if they exist, or just add them.
		for _, command := range includeCommands {
			commands[command.Name] = command
		}
	}

    // Get the name of all the commands (the keys) so we can sort them,
    // otherwise hash maps aren't sorted and would output in a random order.
	var names []string
	for k := range commands {
		names = append(names, k)
	}
	sort.Strings(names)
	// Iternate over the sorted name list so that the final subcommand array,
	// which are now in alphabetical order.
	for _, name := range names {
		subCommands = append(subCommands, commands[name])
	}
	return subCommands
}

// TODO
func getCommands(config Config) []cli.Command {
	exportCmds := []cli.Command{}

    // Sort the list of commands so they show up in alphabetical order.
	var keys []string
	for k := range config.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

    // For each command in the config, create a proper cli.Command{} struct for it.
	for _, name := range keys {
		// Pull out the command we're working on.
		cmd := config.Commands[name]
		// And keep track of the command name.
		cmdName := name

        // Create a brand new cli.Command struct to hold the details from the config's command.
		newCmd := cli.Command{
			Name:            name,
			// Don't parse flags after 'ahoy'.. so 'ahoy -v' works, but 
			// 'ahoy somecmd -v' doesn't.. -v is passed to subcommand.
			// This avoids the need double dashed like 'ahoy somecmd -- -a-flag'
			SkipFlagParsing: true,

			// Don't show the help text or even the command itself in the list if
			// 'hide: true' is set.
			HideHelp:        cmd.Hide,
		}
        
        // Set a usage if one was set in the config's command.
		if cmd.Usage != "" {
			newCmd.Usage = cmd.Usage
		}

        // Set the action to do (the actual bash to run), if there was one set.
		if cmd.Cmd != "" {
			newCmd.Action = func(c *cli.Context) {
				args = c.Args()
				runCommand(cmdName, cmd.Cmd)
			}
		}
        // Load all the subcommands that were set using 'imports: []' in the config.
		subCommands := getSubCommands(cmd.Imports)
		if subCommands != nil {
			newCmd.Subcommands = subCommands
		}

		//log.Println("found command: ", name, " > ", cmd.Cmd )
		// Finally add the newly created commands to the list.
		exportCmds = append(exportCmds, newCmd)
	}

	return exportCmds
}

// TODO
// This actually executes the command that was called.
func runCommand(name string, c string) {

    // Find and replace any use of {{args}} with the actual stuff that was passed to the ahoy command.
    // Example: 'ahoy mycmd some paramters' would replace {{args}} with 'some parameters' if set.
	cReplace := strings.Replace(c, "{{args}}", strings.Join(args, " "), -1)

    // Keep track of the directory where the .ahoy.yml file was since we need to use that as a base.
	dir := sourcedir

    // If the -v flag is set, then first output what bash script code we're actually going 
    // to run (after replacements). Helps with debugging.
	if verbose {
		log.Println("===> AHOY", name, "from", sourcefile, ":", cReplace)
	}

	// Configure the bash script to execute using `bash -c '#myscript'` 
	cmd := exec.Command("bash", "-c", cReplace)
	// Use the path of the ahoy.yml file as the base for the command so that there is consistency
	// when you call ahoy from other folders.
	cmd.Dir = dir

	// Connect all Std* so that all the unix file descriptors and pipes work properly.
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// Actually execute the command.
	if err := cmd.Run(); err != nil {
		// If there is an error, output it and exit with an error.
		// TODO: It might make more sense to pass the proper error code and Stderr directly.
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}


//TODO: There are a few commands that are built into ahoy, but they should be able to be
// overwritten.
func addDefaultCommands(commands []cli.Command) []cli.Command {

    // The first default command is 'ahoy init' which should download a sample ahoy file to start with.
	defaultInitCmd := cli.Command{
		Name:  "init",
		Usage: "Initialize a new .ahoy.yml config file in the current directory.",
		Action: func(c *cli.Context) {
			// Grab the URL or use a default for the initial ahoy file.
			// Allows users to define their own files to call to init.
			// TODO: Update this url to the new DevinciHQ url. Also, should we version it?
			var wgetUrl = "https://raw.githubusercontent.com/devinci-code/ahoy/master/examples/examples.ahoy.yml"

			// Allow for `ahoy init http://some-other-yml-file` if someone specifies it.
			if len(c.Args()) > 0 {
				wgetUrl = c.Args()[0]
			}
			grabYaml := "wget " + wgetUrl + " -O .ahoy.yml"
			cmd := exec.Command("bash", "-c", grabYaml)
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr)
				os.Exit(1)
			} else {
				fmt.Println("example.ahoy.yml downloaded to the current directory. You can customize it to suit your needs!")
			}
		},
	}

	// Don't add default commands if they've already been set.
	if c := app.Command(defaultInitCmd.Name); c == nil {
		commands = append(commands, defaultInitCmd)
	}
	return commands
}

//TODO Move these to flag.go?
// Setup some initial flags at init() time, otherwise they don't work?
func init() {
	flag.StringVar(&sourcefile, "f", "", "specify the sourcefile")
	flag.BoolVar(&bashCompletion, "generate-bash-completion", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
}

// Prints the list of subcommands as the default app completion method
func BashComplete(c *cli.Context) {

	if sourcefile != "" {
		log.Println(sourcefile)
		os.Exit(0)
	}
	for _, command := range c.App.Commands {
		for _, name := range command.Names() {
			fmt.Fprintln(c.App.Writer, name)
		}
	}
}

// The main() function is run after init()
func main() {
	initFlags()
	//log.Println(sourcefile)
	// cli stuff
	app = cli.NewApp()
	app.Name = "ahoy"
	app.Usage = "Creates a configurable cli app for running commands."
	app.EnableBashCompletion = true
	app.BashComplete = BashComplete
	overrideFlags(app)
	if sourcefile, err := getConfigPath(sourcefile); err == nil {
		sourcedir = filepath.Dir(sourcefile)
		config, _ := getConfig(sourcefile)
		app.Commands = getCommands(config)
		app.Commands = addDefaultCommands(app.Commands)
		if config.Usage != "" {
			app.Usage = config.Usage
		}
	}
    // This is the template of the help output when typing just `ahoy` or `ahoy -h`
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .Flags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t" }}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`

    // Final call which shows the help or runs the command depending on the situation.
	app.Run(os.Args)
}
