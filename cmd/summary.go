package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/amiraminb/gh-plantir/internal/bedrock"
	"github.com/amiraminb/gh-plantir/internal/config"
	"github.com/amiraminb/gh-plantir/internal/github"
	"github.com/spf13/cobra"
)

var (
	summaryTeamFlag    string
	summaryTimeFlag    string
	summaryProfileFlag string
	summaryRegionFlag  string
	summaryModelFlag   string
	summaryJSONFlag    bool
	summaryListFlag    bool
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarize merged PRs reviewed by a team",
	Long: `Fetches merged pull requests reviewed by a team within a date range,
then uses AWS Bedrock (Claude) to produce a human-readable summary grouped by repository.

Examples:
  gh plantir summary --team kohofinancial/nomads --time 2026-02-11
  gh plantir summary --team kohofinancial/nomads --time 2026-02-01..2026-02-11
  gh plantir summary --team kohofinancial/nomads --time 2026-02-11 --profile my-aws-profile
  gh plantir summary --team kohofinancial/nomads --time 2026-02-11 --json
  gh plantir summary --team kohofinancial/nomads --time 2026-02-11 --list`,
	Run: func(cmd *cobra.Command, args []string) {
		if summaryTeamFlag == "" {
			fmt.Println("Error: --team is required (format: org/team)")
			return
		}
		if summaryTimeFlag == "" {
			fmt.Println("Error: --time is required (format: YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD)")
			return
		}

		startDate, endDate, err := parseTimeFlag(summaryTimeFlag)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Warning: could not load config: %v\n", err)
			cfg = &config.Config{}
		}

		profile := resolve(summaryProfileFlag, cfg.Profile, "")
		region := resolve(summaryRegionFlag, cfg.Region, bedrock.DefaultRegion)
		model := resolve(summaryModelFlag, cfg.Model, bedrock.DefaultModel)

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
			fmt.Println()
			for _, pr := range prs {
				fmt.Printf("  %s/%s #%d  %s  (@%s, merged %s)\n",
					pr.Owner, pr.Repo, pr.Number, pr.Title, pr.Author,
					pr.MergedAt.Format("2006-01-02"))
			}
			return
		}

		if summaryJSONFlag {
			data, err := json.MarshalIndent(prs, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		fmt.Println("Generating summary...")
		client, err := bedrock.NewClient(profile, region, model)
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

func resolve(flag, configVal, defaultVal string) string {
	if flag != "" {
		return flag
	}
	if configVal != "" {
		return configVal
	}
	return defaultVal
}

func init() {
	rootCmd.AddCommand(summaryCmd)

	summaryCmd.Flags().StringVar(&summaryTeamFlag, "team", "", "Team to summarize (format: org/team)")
	summaryCmd.Flags().StringVar(&summaryTimeFlag, "time", "", "Date or range (YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD)")
	summaryCmd.Flags().StringVar(&summaryProfileFlag, "profile", "", "AWS named profile (overrides config)")
	summaryCmd.Flags().StringVar(&summaryRegionFlag, "region", "", "AWS region (overrides config)")
	summaryCmd.Flags().StringVar(&summaryModelFlag, "model", "", "Bedrock model ID (overrides config)")
	summaryCmd.Flags().BoolVar(&summaryJSONFlag, "json", false, "Output raw PR data as JSON instead of AI summary")
	summaryCmd.Flags().BoolVar(&summaryListFlag, "list", false, "List all fetched PRs before summarizing")
}
