package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage back-cal configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a default config file",
	RunE:  runConfigInit,
}

var configInitForce bool

func init() {
	configInitCmd.Flags().BoolVarP(&configInitForce, "force", "f", false, "Overwrite existing config file")
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	dir := configDir()
	path := filepath.Join(dir, "config.yaml")

	if !configInitForce {
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "warn: config file already exists at %s (use --force to overwrite)\n", path)
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("checking config file: %w", err)
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(configTemplate), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Println(path)
	return nil
}

const configTemplate = `# back-cal configuration
# Run 'back-cal setup' after editing this file to register Task Scheduler tasks.

api:
  # Base URL of the radiator frame endpoint
  url: https://gotta-go.notnot.uk/v1/frame

  # x-radiator-slug header — identifies which calendar to fetch
  slug: ""

  # x-radiator-token header — authentication token
  # Prefer setting the BACK_CAL_TOKEN environment variable instead of storing
  # the token in plaintext here.
  token: ""

canvas:
  # Hex colour for the full-screen background canvas (e.g. "#000000" for black)
  background_color: "#000000"

  # Scale factor applied to the calendar image before compositing.
  # 1 = native size as returned by the API, 0.5 = half size, 2 = double size.
  scale: 1

padding:
  # Distance in pixels from the top edge of the screen to the calendar image
  top: 20
  # Distance in pixels from the right edge of the screen to the calendar image
  right: 20

paths:
  # Where the composited wallpaper BMP is written before being set as wallpaper.
  # Leave empty to use the default: %LOCALAPPDATA%\back-cal\wallpaper.bmp
  wallpaper: ""

  # File that records the date the wallpaper was last successfully updated.
  # Leave empty to use the default: %LOCALAPPDATA%\back-cal\last-run.txt
  last_run: ""

  # Append-only log file path.
  # Leave empty to use the default: %LOCALAPPDATA%\back-cal\back-cal.log
  log: ""
`
