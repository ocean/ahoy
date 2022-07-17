![AHOY logo](https://avatars.githubusercontent.com/u/19353604?s=300&v=4)

# AHOY! - Automate and organize your workflows, no matter what technology you use.

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/ahoy-cli/ahoy/tree/master.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/gh/ahoy-cli/ahoy/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/ahoy-cli/ahoy)](https://goreportcard.com/report/github.com/ahoy-cli/ahoy)

### Note: Ahoy 2.x is now released and is the only supported version.

Ahoy is command line tool that gives each of your projects their own CLI app with zero code and dependencies.

Write your commands in a YAML file and then Ahoy gives you lots of features like:
* a command listing
* per-command help text
* command tab completion
* run commands from any subdirectory

Ahoy makes it easy to create aliases and templates for commands that are useful. It was created to help with running interactive commands within Docker containers, but it's just as useful for local commands, commands over `ssh`, or really anything that could be run from the command line in a single clean interface.

## Examples

Say you want to import a MySQL database running in `docker-compose` using another container called `cli`. The command could look like this:

`docker exec -i $(docker-compose ps -q cli) bash -c 'mysql -u$DB_ENV_MYSQL_USER -p$DB_ENV_MYSQL_PASSWORD -h$DB_PORT_3306_TCP_ADDR $DB_ENV_MYSQL_DATABASE' < some-database.sql`

With Ahoy, you can turn this into:

`ahoy mysql-import < some-database.sql`

[More examples](Home.html).

## Features

- Non-invasive - Use your existing workflow! It can wrap commands and scripts you are already using.
- Consistent - Commands always run relative to the `.ahoy.yml` file, but can be called from any subfolder.
- Visual - See a list of all your commands in one place, along with helpful descriptions.
- Flexible - Commands are specific to a single folder tree, so each repo/workspace can have its own commands.
- Command templates - Args can be dropped into your commands using `{{args}}`
- Fully interactive - Your shells (like MySQL) and prompts still work.
- Self-documenting - Commands and help declared in `.ahoy.yml` show up as ahoy command help and shell completion of commands (see [bash/zsh completion](#bash-zsh-completion)) is also available.

## Installation

### macOS

Using Homebrew:

```
brew install ahoy
```

Note that `ahoy` is in `homebrew-core` as of 1/18/19, so you don't need to use the old tap.
If you were previously using it, you can use the following command to remove it:

```
brew untap ahoy-cli/tap
```

### Linux

Download the [latest release from GitHub](https://github.com/ahoy-cli/ahoy/releases), move the appropriate binary for your plaform into someplace in your $PATH and rename it `ahoy`.

Example:
```
os=$(uname -s | tr [:upper:] [:lower:]) && architecture=$(case $(uname -m) in x86_64 | amd64) echo "amd64" ;; aarch64 | arm64 | armv8) echo "arm64" ;; *) echo "amd64" ;; esac) && sudo wget -q https://github.com/ahoy-cli/ahoy/releases/download/2.0.1/ahoy-bin-$os-$architecture -O /usr/local/bin/ahoy && sudo chown $USER /usr/local/bin/ahoy && chmod +x /usr/local/bin/ahoy
```

### Windows

For WSL2, use the Linux binary above for your architecture.

## Shell Completion (for Bash / Zsh currently)

For Zsh, Just add this to your `~/.zshrc`, and your completions will be relative to the directory you're in.

`complete -F "ahoy --generate-bash-completion" ahoy`

For Bash, you'll need to make sure you have `bash-completion` installed and setup. On macOS with `homebrew` it looks like this:

`brew install bash bash-completion`

Now make sure you follow the installation instructions in the "Caveats" section that `homebrew` returns. And make sure completion is working for something, e.g. `git`, before you continue (you may need to restart your shell).

Then, (for `homebrew`) you'll want to create a file at `/usr/local/etc/bash_completion.d/ahoy` with the following:

```Bash
#! /bin/bash

: ${PROG:=$(basename ${BASH_SOURCE})}

_cli_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

 complete -F _cli_bash_autocomplete $PROG
```

Restart your shell, and you should see `ahoy` autocomplete when typing `ahoy [TAB]`

## Usage

Almost all the commands are actually specified in an `.ahoy.yml` file placed in your working tree somewhere. Commands that are added there show up as options in ahoy. Here is what it looks like when using the [example.ahoy.yml file](https://github.com/ahoy-cli/ahoy/blob/master/examples/examples.ahoy.yml). To start with this file locally you can run `ahoy init`.

```
$ ahoy
NAME:
   ahoy - Send commands to docker-compose services

USAGE:
   ahoy [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   vdown  Stop the vagrant box if one exists.
   vup    Start the vagrant box if one exists.
   start  Start the docker compose-containers.
   stop   Stop the docker-compose containers.
   restart  Restart the docker-compose containers.
   drush  Run drush commands in the cli service container.
   bash   Start a shell in the container (like ssh without actual ssh).
   sqlc   Connect to the default mysql database. Supports piping of data into the command.
   behat  Run the behat tests within the container.
   ps   List the running docker-compose containers.
   behat-init Use composer to install behat dependencies.
   init   Initialize a new .ahoy.yml config file in the current directory.
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --generate-bash-completion
   --version, -v    print the version
```

### New Features in v2
- Implements a new feature to import multiple config files using the "imports" field.
- Uses the "last in wins" rule to deal with duplicate commands amongst the config files.
- Better handling of quotes by no longer using `{{args}}`. Use regular bash syntax like `"$@"` for all arguments, or `$1` for the first argument.
- You can now use a different entrypoint (the thing that runs your commands) instead of bash. Ex. using PHP, Node.js, Python, etc.
- Plugins are now possible by overriding the entrypoint.

```
commands:
  list:
      usage: List the commands from the imported config files.
      imports:
        - ./confirmation.ahoy.yml
        - ./docker.ahoy.yml
        - ./examples.ahoy.yml
```

### Planned v2 features
- Provide "drivers" or "plugins" for bash, docker-compose, kubernetes (these systems still work now, this would just make it easier)
- Do specific arg replacement like {{arg1}} and enable specifying specific arguments and flags in the ahoy file itself to cut down on parsing arguments in scripts.
- Support for more built-in commands or a "verify" yaml option that would create a yes / no prompt for potentially destructive commands. (Are you sure you want to delete all your containers?)
- Pipe tab completion to another command (allows you to get tab completion)
