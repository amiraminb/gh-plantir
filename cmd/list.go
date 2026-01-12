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
	repoFlag     string
	jsonFlag     bool
	limitFlag    int
	reviewedFlag bool
	pendingFlag  bool
	teamFlag     string
	mentionsFlag bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List PRs related to your reviews",
	Long:  `Fetches all open pull requests where you are requested as reviewer or have reviewed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var prs []github.PR
		var err error
		var emptyMsg, headerMsg string

		if reviewedFlag && pendingFlag {
			fmt.Println("Error: cannot use both --reviewed and --pending")
			return
		}

		if teamFlag != "" && pendingFlag {
			prs, err = github.FetchTeamReviewRequests(teamFlag)
			emptyMsg = fmt.Sprintf("✨ No PRs waiting for team %s!", teamFlag)
			headerMsg = fmt.Sprintf("👥 PRs waiting for team %s...", teamFlag)
		} else if teamFlag != "" {
			prs, err = github.FetchTeamAll(teamFlag)
			emptyMsg = fmt.Sprintf("✨ No PRs for team %s!", teamFlag)
			headerMsg = fmt.Sprintf("👥 All PRs for team %s (pending + reviewed)...", teamFlag)
		} else if mentionsFlag {
			prs, err = github.FetchMentions()
			emptyMsg = "✨ No PRs where you're mentioned!"
			headerMsg = "💬 PRs where you're mentioned or commented..."
		} else if reviewedFlag {
			prs, err = github.FetchReviewed()
			emptyMsg = "✨ No PRs you've reviewed!"
			headerMsg = "👀 PRs you've reviewed..."
		} else if pendingFlag {
			prs, err = github.FetchReviewRequests()
			emptyMsg = "✨ No PRs waiting for your review!"
			headerMsg = "🔍 PRs waiting for your review..."
		} else {
			prs, err = github.FetchAll()
			emptyMsg = "✨ No PRs related to you!"
			headerMsg = "🔮 All PRs (pending + reviewed + mentioned)..."
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		prs = filter.Apply(prs, filter.Options{
			Repo: repoFlag,
		})

		sort.Slice(prs, func(i, j int) bool {
			// Primary sort: by age (newest first)
			if !prs[i].CreatedAt.Equal(prs[j].CreatedAt) {
				return prs[i].CreatedAt.After(prs[j].CreatedAt)
			}
			// Secondary sort: features before dependabot
			return prs[i].Type() == "feature" && prs[j].Type() != "feature"
		})

		totalCount := len(prs)

		if totalCount == 0 {
			if jsonFlag {
				fmt.Println("[]")
				return
			}
			fmt.Println(emptyMsg)
			return
		}

		hasMore := false
		if limitFlag > 0 && len(prs) > limitFlag {
			prs = prs[:limitFlag]
			hasMore = true
		}

		if jsonFlag {
			output.JSON(prs)
		} else {
			fmt.Println(headerMsg)
			if hasMore {
				fmt.Printf("\nShowing %d of %d PRs (use --limit to see more):\n\n", limitFlag, totalCount)
			} else {
				fmt.Printf("\nFound %d PRs:\n\n", totalCount)
			}
			output.Table(prs)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&repoFlag, "repo", "", "Filter by repository name")
	listCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	listCmd.Flags().IntVarP(&limitFlag, "limit", "n", 20, "Maximum number of PRs to show (0 for unlimited)")
	listCmd.Flags().BoolVarP(&pendingFlag, "pending", "p", false, "Show only PRs waiting for your review")
	listCmd.Flags().BoolVarP(&reviewedFlag, "reviewed", "r", false, "Show only PRs you've already reviewed")
	listCmd.Flags().BoolVarP(&mentionsFlag, "mentions", "m", false, "Show PRs where you're mentioned or commented")
	listCmd.Flags().StringVar(&teamFlag, "team", "", "Show PRs for a team (format: org/team). Use with -p for pending only")
}
