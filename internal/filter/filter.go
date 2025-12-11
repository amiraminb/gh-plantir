package filter

import (
	"strings"
	"time"

	"github.com/amiraminb/plantir/internal/github"
)

type Options struct {
	Repo      string
	Type      string
	StaleDays int
}

func Apply(prs []github.PR, opts Options) []github.PR {
	var result []github.PR

	for _, pr := range prs {
		if opts.Repo != "" && !strings.Contains(strings.ToLower(pr.Repo), strings.ToLower(opts.Repo)) {
			continue
		}

		if opts.Type != "" {
			if opts.Type == "feature" && pr.Type() != "feature" {
				continue
			}
			if opts.Type == "dependabot" && pr.Type() != "dependabot" {
				continue
			}
			if opts.Type == "bot" && pr.Type() == "feature" {
				continue
			}
		}

		if opts.StaleDays > 0 {
			age := time.Since(pr.CreatedAt)
			if age.Hours() < float64(opts.StaleDays*24) {
				continue
			}
		}

		result = append(result, pr)
	}

	return result
}
