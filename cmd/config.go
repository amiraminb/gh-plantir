package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/amiraminb/gh-plantir/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure default settings for gh-plantir",
	Long:  "Interactively set default AWS profile, region, and Bedrock model for the summary command.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		scanner := bufio.NewScanner(os.Stdin)

		cfg.Profile = promptField(scanner, "AWS Profile", cfg.Profile, "")
		cfg.Region = promptField(scanner, "AWS Region", cfg.Region, "")
		cfg.Model = promptField(scanner, "Bedrock Model ID", cfg.Model, "")

		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Println("\nConfig saved.")
	},
}

func promptField(scanner *bufio.Scanner, label, current, defaultVal string) string {
	display := current
	if display == "" {
		if defaultVal != "" {
			display = defaultVal + " (default)"
		} else {
			display = "unset"
		}
	}
	fmt.Printf("%s [%s]: ", label, display)

	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			return input
		}
	}

	if current != "" {
		return current
	}
	return ""
}

func init() {
	rootCmd.AddCommand(configCmd)
}
