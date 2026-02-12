package bedrock

import (
	"fmt"
	"sort"
	"strings"

	"github.com/amiraminb/gh-plantir/internal/github"
)

const maxPromptChars = 150000

const systemPrompt = `You are a technical writer summarizing merged pull requests for a software engineering team.

The software engineers you are working for are very professional and busy persons. They care a lot about technical stuff and very passionate to learn technical patterns and best practices. They want to stay on top of what is happening in their team reading codes. Produce a clear, to the point, simple language, concise summary grouped by repository. For each repository, categorize changes into:
- New Features
- Bug Fixes
- Refactors / Improvements
- Other

For each PR, write a one-line summary of what changed and why. Include the PR number and author.
Use markdown formatting.
Produce the takeways from PRs and learnings for best practices. 
`

func BuildPrompt(prs []github.MergedPR) (string, error) {
	// Group PRs by Owner/Repo
	type repoKey struct {
		Owner string
		Repo  string
	}
	grouped := make(map[repoKey][]github.MergedPR)
	var keys []repoKey

	for _, pr := range prs {
		k := repoKey{Owner: pr.Owner, Repo: pr.Repo}
		if _, exists := grouped[k]; !exists {
			keys = append(keys, k)
		}
		grouped[k] = append(grouped[k], pr)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Owner != keys[j].Owner {
			return keys[i].Owner < keys[j].Owner
		}
		return keys[i].Repo < keys[j].Repo
	})

	var b strings.Builder
	b.WriteString("Summarize the following merged pull requests:\n\n")

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("## %s/%s\n\n", k.Owner, k.Repo))
		for _, pr := range grouped[k] {
			body := pr.Body
			if len(body) > 1000 {
				body = body[:1000] + "..."
			}

			b.WriteString(fmt.Sprintf("### PR #%d: %s (by @%s)\n", pr.Number, pr.Title, pr.Author))
			if body != "" {
				b.WriteString(fmt.Sprintf("Description: %s\n", body))
			}
			b.WriteString(fmt.Sprintf("Stats: +%d -%d\n", pr.Additions, pr.Deletions))

			files := pr.Files
			if len(files) > 30 {
				files = files[:30]
			}
			if len(files) > 0 {
				b.WriteString("Changed files:\n")
				for _, f := range files {
					b.WriteString(fmt.Sprintf("  - %s (+%d -%d)\n", f.Path, f.Additions, f.Deletions))
				}
			}
			b.WriteString("\n")
		}
	}

	prompt := b.String()
	if len(prompt) > maxPromptChars {
		return "", fmt.Errorf("prompt too large (%d chars, max %d) — try a narrower date range", len(prompt), maxPromptChars)
	}

	return prompt, nil
}
