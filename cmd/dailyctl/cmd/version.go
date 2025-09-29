package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, build commit, and build date information for dailyctl.`,
	Run: func(cmd *cobra.Command, args []string) {
		v, c, d := GetVersionInfo()
		fmt.Printf("dailyctl version %s\n", v)
		fmt.Printf("Build commit: %s\n", c)
		fmt.Printf("Build date: %s\n", d)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}



