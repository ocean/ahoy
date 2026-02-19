package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// InitArgs contains arguments for the init command.
type InitArgs struct {
	Force bool
	URL   string
}

// downloadFile downloads a file from the given URL and saves it to the specified path.
func downloadFile(url, destPath string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: server returned %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", destPath, err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file %s: %v", destPath, err)
	}

	return nil
}

// RunConfigInit performs the init command functionality.
func RunConfigInit(args InitArgs) error {
	if fileExists(filepath.Join(".", ".ahoy.yml")) {
		if args.Force {
			fmt.Println("Warning: '--force' parameter passed, overwriting .ahoy.yml in current directory.")
		} else {
			fmt.Println("Warning: .ahoy.yml found in current directory.")
			fmt.Fprint(os.Stderr, "Are you sure you wish to overwrite it with an example file, y/N ? ")
			reader := bufio.NewReader(os.Stdin)
			char, _, err := reader.ReadRune()
			if err != nil {
				return fmt.Errorf("failed to read input: %v", err)
			}
			if char != 'y' && char != 'Y' {
				fmt.Println("Abort: exiting without overwriting.")
				return nil
			}
			if args.URL != "" {
				fmt.Println("Ok, overwriting .ahoy.yml in current directory with specified file.")
			} else {
				fmt.Println("Ok, overwriting .ahoy.yml in current directory with example file.")
			}
		}
	}

	downloadURL := "https://raw.githubusercontent.com/ahoy-cli/ahoy/master/examples/examples.ahoy.yml"
	if args.URL != "" {
		downloadURL = args.URL
	}

	if err := downloadFile(downloadURL, ".ahoy.yml"); err != nil {
		return fmt.Errorf("failed to download config file: %v", err)
	}

	if args.URL != "" {
		fmt.Println("Your specified .ahoy.yml has been downloaded to the current directory.")
	} else {
		fmt.Println("Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!")
	}

	return nil
}

// initCommandAction is the Cobra handler for the init command.
func initCommandAction(cmd *cobra.Command, args []string) {
	initArgs := InitArgs{
		Force: func() bool { f, _ := cmd.Flags().GetBool("force"); return f }(),
	}

	if len(args) > 0 {
		initArgs.URL = args[0]
	}

	if err := RunConfigInit(initArgs); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
