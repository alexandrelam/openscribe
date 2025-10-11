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
		fmt.Printf("OpenScribe v%s\n", Version)
		fmt.Printf("Commit:     %s\n", GitCommit)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Platform:   darwin/amd64\n")
		fmt.Printf("\nFor more information, visit: https://github.com/alexandrelam/openscribe-go\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
