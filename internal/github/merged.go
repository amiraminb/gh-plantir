package github

import (
	"fmt"
	"time"
)

type ChangedFile struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type MergedPR struct {
	Number    int           `json:"number"`
	Title     string        `json:"title"`
	URL       string        `json:"url"`
	Body      string        `json:"body"`
	Author    string        `json:"author"`
	Repo      string        `json:"repo"`
	Owner     string        `json:"owner"`
	MergedAt  time.Time     `json:"mergedAt"`
	Labels    []string      `json:"labels"`
	Additions int           `json:"additions"`
	Deletions int           `json:"deletions"`
	Files     []ChangedFile `json:"files"`
}

type mergedSearchResponse struct {
	Search struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []struct {
			Number   int    `json:"number"`
			Title    string `json:"title"`
			URL      string `json:"url"`
			Body     string `json:"body"`
			MergedAt string `json:"mergedAt"`
			Author   struct {
				Login string `json:"login"`
			} `json:"author"`
			Repository struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"repository"`
			Labels struct {
				Nodes []struct {
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"labels"`
			Additions int `json:"additions"`
			Deletions int `json:"deletions"`
			Files     struct {
				Nodes []struct {
					Path      string `json:"path"`
					Additions int    `json:"additions"`
					Deletions int    `json:"deletions"`
				} `json:"nodes"`
			} `json:"files"`
		} `json:"nodes"`
	} `json:"search"`
}

func mergedSearchQuery(searchQuery, cursor string) string {
	after := "null"
	if cursor != "" {
		after = fmt.Sprintf(`"%s"`, cursor)
	}
	return fmt.Sprintf(`
query {
  search(query: "%s", type: ISSUE, first: 100, after: %s) {
    pageInfo {
      hasNextPage
      endCursor
    }
    nodes {
      ... on PullRequest {
        number
        title
        url
        body
        mergedAt
        additions
        deletions
        author { login }
        repository {
          name
          owner { login }
        }
        labels(first: 10) {
          nodes { name }
        }
        files(first: 100) {
          nodes {
            path
            additions
            deletions
          }
        }
      }
    }
  }
}
`, searchQuery, after)
}

func fetchMergedPRs(searchQuery string) ([]MergedPR, error) {
	var all []MergedPR
	cursor := ""

	for {
		query := mergedSearchQuery(searchQuery, cursor)
		var resp mergedSearchResponse
		if err := graphqlQuery(query, &resp); err != nil {
			return nil, fmt.Errorf("failed to query merged PRs: %w", err)
		}

		for _, node := range resp.Search.Nodes {
			labels := make([]string, len(node.Labels.Nodes))
			for i, l := range node.Labels.Nodes {
				labels[i] = l.Name
			}

			files := make([]ChangedFile, len(node.Files.Nodes))
			for i, f := range node.Files.Nodes {
				files[i] = ChangedFile{
					Path:      f.Path,
					Additions: f.Additions,
					Deletions: f.Deletions,
				}
			}

			mergedAt, _ := time.Parse(time.RFC3339, node.MergedAt)

			all = append(all, MergedPR{
				Number:    node.Number,
				Title:     node.Title,
				URL:       node.URL,
				Body:      node.Body,
				Author:    node.Author.Login,
				Repo:      node.Repository.Name,
				Owner:     node.Repository.Owner.Login,
				MergedAt:  mergedAt,
				Labels:    labels,
				Additions: node.Additions,
				Deletions: node.Deletions,
				Files:     files,
			})
		}

		if !resp.Search.PageInfo.HasNextPage {
			break
		}
		cursor = resp.Search.PageInfo.EndCursor
	}

	return all, nil
}

func FetchMergedPRsForTeam(team, startDate, endDate string) ([]MergedPR, error) {
	// Strategy 1: team-review-requested
	teamQuery := fmt.Sprintf(
		"is:pr is:merged merged:%s..%s team-review-requested:%s -author:app/dependabot",
		startDate, endDate, team,
	)
	teamPRs, err := fetchMergedPRs(teamQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch team-requested merged PRs: %w", err)
	}

	// Strategy 2: per-member reviewed-by queries
	members, err := getTeamMembers(team)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	// Strategy 2: per-member reviewed-by queries (one per member — OR doesn't work with these qualifiers)
	var memberPRs []MergedPR
	for _, m := range members {
		memberQuery := fmt.Sprintf(
			"is:pr is:merged merged:%s..%s reviewed-by:%s -author:app/dependabot",
			startDate, endDate, m,
		)
		mPRs, err := fetchMergedPRs(memberQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch reviewed-by %s: %w", m, err)
		}
		memberPRs = append(memberPRs, mPRs...)
	}

	// Deduplicate by URL
	seen := make(map[string]bool)
	var result []MergedPR
	for _, prs := range [][]MergedPR{teamPRs, memberPRs} {
		for _, pr := range prs {
			if !seen[pr.URL] {
				seen[pr.URL] = true
				result = append(result, pr)
			}
		}
	}

	return result, nil
}
