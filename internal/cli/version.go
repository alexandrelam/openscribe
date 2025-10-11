package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information (will be set via build flags)
	Version = "dev"
	// GitCommit is the git commit hash used to build this binary
	GitCommit = "unknown"
	// BuildDate is the date this binary was built
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version, git commit, and build date of OpenScribe.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("OpenScribe %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
