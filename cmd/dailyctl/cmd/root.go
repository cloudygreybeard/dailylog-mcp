package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version string
	commit  string
	date    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dailyctl",
	Short: "Daily Log Control - Command line interface for daily activity logging",
	Long: `dailyctl is a command-line interface for managing daily activity logs.
It provides tools for creating, searching, and summarizing your daily activities,
moods, and notes stored in your private GitHub repository.

Examples:
  dailyctl log activity "Morning meeting with team" --tags work,meeting --mood 8
  dailyctl get today
  dailyctl search --query "exercise" --mood-min 7
  dailyctl summarize week`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(v, c, d string) error {
	version = v
	commit = c
	date = d
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dailyctl.yaml)")
	rootCmd.PersistentFlags().String("github-repo", "", "GitHub repository for storage (owner/repo)")
	rootCmd.PersistentFlags().String("github-token", "", "GitHub personal access token")
	rootCmd.PersistentFlags().String("github-path", "logs", "Path within GitHub repo for logs")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("github.repo", rootCmd.PersistentFlags().Lookup("github-repo"))
	viper.BindPFlag("github.token", rootCmd.PersistentFlags().Lookup("github-token"))
	viper.BindPFlag("github.path", rootCmd.PersistentFlags().Lookup("github-path"))
	viper.BindPFlag("output.format", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".dailyctl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dailyctl")
	}

	// Environment variables
	viper.SetEnvPrefix("DAILYLOG")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verbose") {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// GetVersionInfo returns version information
func GetVersionInfo() (string, string, string) {
	return version, commit, date
}
