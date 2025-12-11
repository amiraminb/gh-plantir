package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/amiraminb/plantir/internal/github"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show summary statistics of your PR reviews",
	Long:  `Displays a breakdown of pending PR reviews by repository and type.`,
	Run: func(cmd *cobra.Command, args []string) {
		prs, err := github.FetchReviewRequests()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(prs) == 0 {
			fmt.Println("✨ No PRs waiting for your review!")
			return
		}

		fmt.Printf("📊 PR Review Stats (%d total)\n\n", len(prs))

		byRepo := make(map[string]int)
		byType := make(map[string]int)
		for _, pr := range prs {
			byRepo[pr.Repo]++
			byType[pr.Type()]++
		}

		fmt.Println("By Repository:")
		repoTable := tablewriter.NewTable(os.Stdout)
		repoTable.Header("Repository", "Count")
		repos := sortedKeys(byRepo)
		for _, repo := range repos {
			repoTable.Append([]string{repo, strconv.Itoa(byRepo[repo])})
		}
		repoTable.Render()

		fmt.Println("\nBy Type:")
		typeTable := tablewriter.NewTable(os.Stdout)
		typeTable.Header("Type", "Count")
		types := sortedKeys(byType)
		for _, t := range types {
			typeTable.Append([]string{t, strconv.Itoa(byType[t])})
		}
		typeTable.Render()
	},
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	return keys
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
