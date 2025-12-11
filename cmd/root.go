package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "plantir",
	Short: "🔮 The seeing stone for your PR reviews",
	Long:  "🔮 Plantir helps you manage GitHub pull requests where you're requested as a reviewer.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
