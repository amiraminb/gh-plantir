package filter

import (
	"strings"

	"github.com/amiraminb/gh-plantir/internal/github"
)

type Options struct {
	Repo          string
	ExcludeDrafts bool
}

func Apply(prs []github.PR, opts Options) []github.PR {
	var result []github.PR

	for _, pr := range prs {
		if opts.Repo != "" && !strings.Contains(strings.ToLower(pr.Repo), strings.ToLower(opts.Repo)) {
			continue
		}

		if opts.ExcludeDrafts && pr.IsDraft {
			continue
		}

		result = append(result, pr)
	}

	return result
}
