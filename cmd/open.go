package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/amiraminb/gh-plantir/internal/github"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/spf13/cobra"
)

var openTeamFlag string

var openCmd = &cobra.Command{
	Use:   "open <PR#>",
	Short: "Open a PR in your browser",
	Long:  `Opens the specified pull request in your default browser.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prNumber, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error: '%s' is not a valid PR number\n", args[0])
			return
		}

		var prs []github.PR
		if openTeamFlag != "" {
			prs, err = github.FetchTeamAll(openTeamFlag)
		} else {
			prs, err = github.FetchAll()
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		for _, pr := range prs {
			if pr.Number == prNumber {
				fmt.Printf("Opening %s#%d in browser...\n", pr.Repo, pr.Number)
				b := browser.New("", os.Stdout, os.Stderr)
				if err := b.Browse(pr.URL); err != nil {
					fmt.Printf("Error opening browser: %v\n", err)
				}
				return
			}
		}

		fmt.Printf("PR #%d not found in your PRs\n", prNumber)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)

	openCmd.Flags().StringVarP(&openTeamFlag, "team", "t", "", "Search within a team's PRs (format: org/team)")
}
