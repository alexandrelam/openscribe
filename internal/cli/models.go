package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Model management",
	Long:  `List and download Whisper models.`,
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and downloaded models",
	Long:  `Display all available Whisper models and indicate which are downloaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Models list command - Not yet implemented")
		fmt.Println("This will list available Whisper models")
	},
}

var modelsDownloadCmd = &cobra.Command{
	Use:   "download [model]",
	Short: "Download a specific Whisper model",
	Long:  `Download a specific Whisper model (tiny, base, small, medium, large).`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a model to download (tiny, base, small, medium, large)")
			return
		}
		fmt.Printf("Models download command - Not yet implemented\n")
		fmt.Printf("This will download the %s model\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsDownloadCmd)
}
