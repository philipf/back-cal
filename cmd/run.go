package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/philipf/back-cal/internal/composite"
	"github.com/philipf/back-cal/internal/fetch"
	"github.com/philipf/back-cal/internal/monitor"
	"github.com/philipf/back-cal/internal/wallpaper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runForce bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch calendar and set as desktop wallpaper",
	RunE:  runRun,
}

func init() {
	runCmd.Flags().BoolVarP(&runForce, "force", "f", false, "Bypass the datestamp guard and run regardless")
	rootCmd.AddCommand(runCmd)
	// Make run the default when no subcommand is given
	rootCmd.RunE = runCmd.RunE
}

func runRun(cmd *cobra.Command, args []string) error {
	logger, err := openLogger(resolvedPath("paths.log", "back-cal.log"))
	if err != nil {
		// Fall back to stderr if we can't open the log file
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	logger.Printf("INFO  run started (force=%v)", runForce)

	// Datestamp guard
	guardPath := resolvedPath("paths.last_run", "last-run.txt")
	if !runForce {
		if ranToday(guardPath) {
			logger.Printf("INFO  already ran today, exiting")
			return nil
		}
	}

	// Validate required config
	slug := viper.GetString("api.slug")
	token := viper.GetString("api.token")
	if slug == "" {
		return logAndExit(logger, "api.slug is required — run 'back-cal config init' to create a config file")
	}
	if token == "" {
		return logAndExit(logger, "api.token is required — set BACK_CAL_TOKEN or add token to config")
	}

	// Detect primary monitor
	width, height, err := monitor.PrimaryMonitor()
	if err != nil {
		logger.Printf("ERROR detecting primary monitor: %v", err)
		return fmt.Errorf("detecting primary monitor: %w", err)
	}
	logger.Printf("INFO  primary monitor: %dx%d", width, height)

	// Fetch BMP
	bmpData, err := fetch.BMP(fetch.Config{
		URL:   viper.GetString("api.url"),
		Slug:  slug,
		Token: token,
	})
	if err != nil {
		logger.Printf("ERROR fetching calendar: %v", err)
		return nil // soft exit — retry next trigger
	}
	logger.Printf("INFO  fetched %d bytes", len(bmpData))

	// Decode BMP
	src, err := composite.DecodeBMP(bmpData)
	if err != nil {
		logger.Printf("ERROR decoding BMP: %v", err)
		return nil
	}
	logger.Printf("INFO  calendar size: %dx%d", src.Bounds().Dx(), src.Bounds().Dy())

	// Composite
	wallpaperPath := resolvedPath("paths.wallpaper", "wallpaper.png")
	err = composite.Composite(src, composite.Config{
		MonitorWidth:  width,
		MonitorHeight: height,
		BgColor:       viper.GetString("canvas.background_color"),
		PaddingTop:    viper.GetInt("padding.top"),
		PaddingRight:  viper.GetInt("padding.right"),
		OutputPath:    wallpaperPath,
	})
	if err != nil {
		logger.Printf("ERROR compositing: %v", err)
		return nil
	}
	logger.Printf("INFO  composited wallpaper written to %s", wallpaperPath)

	// Set wallpaper
	if err := wallpaper.Set(wallpaperPath); err != nil {
		logger.Printf("ERROR setting wallpaper: %v", err)
		return nil
	}
	logger.Printf("INFO  wallpaper set")

	// Update datestamp on success
	if err := writeToday(guardPath); err != nil {
		logger.Printf("WARN  could not update last-run file: %v", err)
	}

	logger.Printf("INFO  done")
	return nil
}

// resolvedPath returns the configured path for key, falling back to a filename
// in the platform-appropriate local data directory.
func resolvedPath(key, filename string) string {
	if p := viper.GetString(key); p != "" {
		return p
	}
	return filepath.Join(localAppDataDir(), filename)
}

func openLogger(path string) (*log.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return log.New(f, "", log.LstdFlags), nil
}

func ranToday(path string) bool {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		return false
	}
	return string(data) == today()
}

func writeToday(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(today()), 0o644)
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func logAndExit(logger *log.Logger, msg string) error {
	logger.Printf("ERROR %s", msg)
	return fmt.Errorf("%s", msg)
}
