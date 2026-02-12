package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/amiraminb/gh-plantir/internal/bedrock"
	"github.com/amiraminb/gh-plantir/internal/config"
	"github.com/amiraminb/gh-plantir/internal/github"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	summaryTeamFlag string
	summaryTimeFlag string
	summaryListFlag bool
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarize merged PRs reviewed by a team",
	Long: `Fetches merged pull requests reviewed by a team within a date range,
then uses AWS Bedrock (Claude) to produce a human-readable summary grouped by repository.

Examples:
  gh plantir summary --team kohofinancial/nomads --date 2026-02-11
  gh plantir summary --team kohofinancial/nomads --date 2026-02-01..2026-02-11
  gh plantir summary --team kohofinancial/nomads --date 2026-02-11 --list`,
	Run: func(cmd *cobra.Command, args []string) {
		if summaryTeamFlag == "" {
			fmt.Println("Error: --team is required (format: org/team)")
			return
		}
		if summaryTimeFlag == "" {
			fmt.Println("Error: --date is required (format: YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD)")
			return
		}

		startDate, endDate, err := parseTimeFlag(summaryTimeFlag)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Fetching merged PRs for team %s (%s to %s)...\n", summaryTeamFlag, startDate, endDate)
		prs, err := github.FetchMergedPRsForTeam(summaryTeamFlag, startDate, endDate)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Found %d merged PRs.\n", len(prs))
		if len(prs) == 0 {
			return
		}

		if summaryListFlag {
			prNum := color.New(color.FgCyan).SprintFunc()
			prTitle := color.New(color.FgWhite).SprintFunc()
			prAuthor := color.New(color.FgMagenta).SprintFunc()
			repoName := color.New(color.FgYellow, color.Bold).SprintFunc()

			grouped := make(map[string][]github.MergedPR)
			var repos []string
			for _, pr := range prs {
				key := pr.Owner + "/" + pr.Repo
				if _, exists := grouped[key]; !exists {
					repos = append(repos, key)
				}
				grouped[key] = append(grouped[key], pr)
			}
			for _, repo := range repos {
				fmt.Printf("\n%s\n", repoName(repo))
				for _, pr := range grouped[repo] {
					fmt.Printf("  %s  %s  %s\n",
						prNum(fmt.Sprintf("#%-6d", pr.Number)),
						prTitle(pr.Title),
						prAuthor(fmt.Sprintf("(@%s)", pr.Author)))
				}
			}
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		if cfg.Profile == "" || cfg.Region == "" || cfg.Model == "" {
			fmt.Println("AWS Bedrock is not fully configured. Run 'gh plantir config' to set your profile, region, and model.")
			return
		}

		fmt.Println("Generating summary...")
		client, err := bedrock.NewClient(cfg.Profile, cfg.Region, cfg.Model)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		summary, err := client.Summarize(prs)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println()
		fmt.Println(summary)
	},
}

func parseTimeFlag(value string) (string, string, error) {
	if strings.Contains(value, "..") {
		parts := strings.SplitN(value, "..", 2)
		start, end := parts[0], parts[1]
		if err := validateDate(start); err != nil {
			return "", "", fmt.Errorf("invalid start date: %w", err)
		}
		if err := validateDate(end); err != nil {
			return "", "", fmt.Errorf("invalid end date: %w", err)
		}
		s, _ := time.Parse("2006-01-02", start)
		e, _ := time.Parse("2006-01-02", end)
		if s.After(e) {
			return "", "", fmt.Errorf("start date %s is after end date %s", start, end)
		}
		return start, end, nil
	}

	if err := validateDate(value); err != nil {
		return "", "", err
	}
	return value, value, nil
}

func validateDate(s string) error {
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date %q (expected YYYY-MM-DD)", s)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(summaryCmd)

	summaryCmd.Flags().StringVarP(&summaryTeamFlag, "team", "t", "", "Team to summarize (format: org/team)")
	summaryCmd.Flags().StringVarP(&summaryTimeFlag, "date", "d", "", "Date or range (YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD)")
	summaryCmd.Flags().BoolVarP(&summaryListFlag, "list", "l", false, "List all fetched PRs instead of summarizing")
}
