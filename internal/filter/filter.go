package filter

import (
	"strings"

	"github.com/amiraminb/plantir/internal/github"
)

type Options struct {
	Repo string
}

func Apply(prs []github.PR, opts Options) []github.PR {
	var result []github.PR

	for _, pr := range prs {
		if opts.Repo != "" && !strings.Contains(strings.ToLower(pr.Repo), strings.ToLower(opts.Repo)) {
			continue
		}

		result = append(result, pr)
	}

	return result
}
