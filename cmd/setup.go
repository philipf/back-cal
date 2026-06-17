package cmd

import (
	"fmt"
	"os"

	"github.com/philipf/back-cal/internal/scheduler"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Register Task Scheduler tasks for automatic daily wallpaper updates",
	Long: `Registers two Task Scheduler tasks under the current user (no elevation required):
  back-cal-logon  — fires on user logon
  back-cal-unlock — fires on workstation unlock (Event ID 4801)

Run this once after installation. Re-running overwrites existing tasks.`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if err := scheduler.Register(execPath); err != nil {
		return err
	}

	fmt.Println("setup complete — tasks will trigger back-cal run on logon and unlock")
	return nil
}
