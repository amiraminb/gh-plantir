package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/amiraminb/plantir/internal/ai"
	"github.com/amiraminb/plantir/internal/github"
	"github.com/spf13/cobra"
)

var (
	reviewRepoFlag    string
	reviewProfileFlag string
	reviewRegionFlag  string
	reviewModelFlag   string
)

type prInfo struct {
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Body      string   `json:"body"`
	Author    author   `json:"author"`
	Labels    []label  `json:"labels"`
	Additions int      `json:"additions"`
	Deletions int      `json:"deletions"`
	Files     []file   `json:"files"`
	BaseRef   string   `json:"baseRefName"`
	HeadRef   string   `json:"headRefName"`
}

type author struct {
	Login string `json:"login"`
}

type label struct {
	Name string `json:"name"`
}

type file struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

var reviewCmd = &cobra.Command{
	Use:   "review <PR#>",
	Short: "Get an AI-powered review of a PR",
	Long:  `Fetches the PR diff and uses Claude to provide a summary and code review.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prNumber, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error: '%s' is not a valid PR number\n", args[0])
			return
		}

		// If no repo specified, try to find it from user's PR list
		repo := reviewRepoFlag
		if repo == "" {
			prs, err := github.FetchAll()
			if err == nil {
				for _, pr := range prs {
					if pr.Number == prNumber {
						repo = pr.Owner + "/" + pr.Repo
						break
					}
				}
			}
		}

		if repo == "" {
			fmt.Printf("Error: PR #%d not found in your PRs. Specify repo with --repo=owner/repo\n", prNumber)
			return
		}

		fmt.Printf("Fetching PR #%d from %s...\n", prNumber, repo)

		// Fetch PR info
		info, err := fetchPRInfo(prNumber, repo)
		if err != nil {
			fmt.Printf("Error fetching PR info: %v\n", err)
			return
		}

		// Fetch PR diff
		diff, err := fetchPRDiff(prNumber, repo)
		if err != nil {
			fmt.Printf("Error fetching PR diff: %v\n", err)
			return
		}

		if len(diff) == 0 {
			fmt.Println("PR has no changes")
			return
		}

		// Truncate diff if too large (Claude has context limits)
		maxDiffSize := 100000 // ~100KB
		if len(diff) > maxDiffSize {
			diff = diff[:maxDiffSize] + "\n\n... [diff truncated due to size]"
			fmt.Println("Note: Diff was truncated due to size")
		}

		fmt.Println("Analyzing with Claude...")

		// Create AI client
		client, err := ai.NewClient(reviewProfileFlag, reviewRegionFlag, reviewModelFlag)
		if err != nil {
			fmt.Printf("Error creating AI client: %v\n", err)
			fmt.Println("\nMake sure you're logged in: aws sso login --profile", reviewProfileFlag)
			return
		}

		// Build PR context
		var labels []string
		for _, l := range info.Labels {
			labels = append(labels, l.Name)
		}

		var files []string
		for _, f := range info.Files {
			files = append(files, f.Path)
		}

		prCtx := ai.PRContext{
			Title:       info.Title,
			URL:         info.URL,
			Description: info.Body,
			Author:      info.Author.Login,
			Labels:      labels,
			Additions:   info.Additions,
			Deletions:   info.Deletions,
			Files:       files,
			BaseBranch:  info.BaseRef,
			HeadBranch:  info.HeadRef,
			Diff:        diff,
		}

		// Get review
		result, err := client.ReviewPR(prCtx)
		if err != nil {
			fmt.Printf("Error getting review: %v\n", err)
			return
		}

		fmt.Println()
		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Printf("PR #%d: %s\n", prNumber, info.Title)
		fmt.Println(info.URL)
		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Println()
		fmt.Println(result.Summary)
	},
}

func fetchPRInfo(prNumber int, repo string) (*prInfo, error) {
	fields := "title,url,body,author,labels,additions,deletions,files,baseRefName,headRefName"
	args := []string{"pr", "view", strconv.Itoa(prNumber), "--json", fields}
	if repo != "" {
		args = append(args, "--repo", repo)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info prInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func fetchPRDiff(prNumber int, repo string) (string, error) {
	args := []string{"pr", "diff", strconv.Itoa(prNumber)}
	if repo != "" {
		args = append(args, "--repo", repo)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func init() {
	rootCmd.AddCommand(reviewCmd)

	reviewCmd.Flags().StringVarP(&reviewRepoFlag, "repo", "r", "", "Repository (owner/repo format)")
	reviewCmd.Flags().StringVar(&reviewProfileFlag, "profile", ai.DefaultProfile, "AWS profile for Bedrock")
	reviewCmd.Flags().StringVar(&reviewRegionFlag, "region", ai.DefaultRegion, "AWS region for Bedrock")
	reviewCmd.Flags().StringVar(&reviewModelFlag, "model", ai.DefaultModel, "Claude model ID")
}
