package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

const reviewRequestQuery = `
query {
  search(query: "is:pr is:open review-requested:@me", type: ISSUE, first: 100) {
    nodes {
      ... on PullRequest {
        number
        title
        url
        isDraft
        createdAt
        author { login }
        repository {
          name
          owner { login }
        }
        labels(first: 10) {
          nodes { name }
        }
      }
    }
  }
}
`

const mentionsQuery = `
query {
  search(query: "is:pr is:open (mentions:@me OR commenter:@me) -author:@me", type: ISSUE, first: 100) {
    nodes {
      ... on PullRequest {
        number
        title
        url
        isDraft
        createdAt
        author { login }
        repository {
          name
          owner { login }
        }
        labels(first: 10) {
          nodes { name }
        }
      }
    }
  }
}
`

type graphQLResponse struct {
	Data struct {
		Search struct {
			Nodes []struct {
				Number    int    `json:"number"`
				Title     string `json:"title"`
				URL       string `json:"url"`
				IsDraft   bool   `json:"isDraft"`
				CreatedAt string `json:"createdAt"`
				Author    struct {
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
			} `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

func FetchReviewRequests() ([]PR, error) {
	return fetchPRs(reviewRequestQuery)
}

func FetchMentions() ([]PR, error) {
	return fetchPRs(mentionsQuery)
}

func fetchPRs(query string) ([]PR, error) {
	cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", query))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to call gh api: %w", err)
	}

	var resp graphQLResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var prs []PR
	for _, node := range resp.Data.Search.Nodes {
		labels := make([]string, len(node.Labels.Nodes))
		for i, l := range node.Labels.Nodes {
			labels[i] = l.Name
		}

		createdAt, _ := time.Parse(time.RFC3339, node.CreatedAt)

		prs = append(prs, PR{
			Number:    node.Number,
			Title:     node.Title,
			URL:       node.URL,
			Author:    node.Author.Login,
			Repo:      node.Repository.Name,
			Owner:     node.Repository.Owner.Login,
			IsDraft:   node.IsDraft,
			Labels:    labels,
			CreatedAt: createdAt,
		})
	}

	return prs, nil
}
