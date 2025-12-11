package cmd

import (
	"fmt"
	"sort"

	"github.com/amiraminb/plantir/internal/filter"
	"github.com/amiraminb/plantir/internal/github"
	"github.com/amiraminb/plantir/internal/output"
	"github.com/spf13/cobra"
)

var (
	mentionsRepoFlag  string
	mentionsLimitFlag int
	mentionsJsonFlag  bool
)

var mentionsCmd = &cobra.Command{
	Use:   "mentions",
	Short: "List PRs where you're mentioned or involved in discussions",
	Long:  `Fetches open PRs where you're mentioned in comments or have participated in discussions.`,
	Run: func(cmd *cobra.Command, args []string) {
		prs, err := github.FetchMentions()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		prs = filter.Apply(prs, filter.Options{
			Repo: mentionsRepoFlag,
		})

		sort.Slice(prs, func(i, j int) bool {
			return prs[i].CreatedAt.After(prs[j].CreatedAt)
		})

		totalCount := len(prs)

		if totalCount == 0 {
			if mentionsJsonFlag {
				fmt.Println("[]")
				return
			}
			fmt.Println("✨ No PRs with mentions or discussions!")
			return
		}

		hasMore := false
		if mentionsLimitFlag > 0 && len(prs) > mentionsLimitFlag {
			prs = prs[:mentionsLimitFlag]
			hasMore = true
		}

		if mentionsJsonFlag {
			output.JSON(prs)
		} else {
			fmt.Println("💬 PRs where you're mentioned or involved...")
			if hasMore {
				fmt.Printf("\nShowing %d of %d PRs (use --limit to see more):\n\n", mentionsLimitFlag, totalCount)
			} else {
				fmt.Printf("\nFound %d PRs:\n\n", totalCount)
			}
			output.Table(prs)
		}
	},
}

func init() {
	rootCmd.AddCommand(mentionsCmd)

	mentionsCmd.Flags().StringVarP(&mentionsRepoFlag, "repo", "r", "", "Filter by repository name")
	mentionsCmd.Flags().IntVarP(&mentionsLimitFlag, "limit", "n", 20, "Maximum number of PRs to show (0 for unlimited)")
	mentionsCmd.Flags().BoolVar(&mentionsJsonFlag, "json", false, "Output as JSON")
}
