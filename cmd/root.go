package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "back-cal",
	Short: "Fetch and display a calendar as the desktop wallpaper",
	Long:  "back-cal fetches a calendar image from a remote API and composites it onto the desktop wallpaper, positioned top-right.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetEnvPrefix("BACK_CAL")
	viper.AutomaticEnv()
	viper.BindEnv("api.token", "BACK_CAL_TOKEN")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir())

	viper.SetDefault("api.url", "https://gotta-go.notnot.uk/v1/frame")
	viper.SetDefault("canvas.background_color", "#000000")
	viper.SetDefault("padding.top", 20)
	viper.SetDefault("padding.right", 20)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "warn: could not read config: %v\n", err)
		}
	}
}

func configDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "back-cal")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "back-cal")
}

func localAppDataDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "back-cal")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "back-cal")
}
