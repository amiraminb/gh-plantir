package cmd

import (
	"sort"
	"strings"

	"github.com/amiraminb/gh-plantir/internal/github"
)

type listMode int

const (
	listModeMixed listMode = iota
	listModePending
	listModeReviewed
	listModeMentions
)

func currentListMode(pending, reviewed, mentions bool) listMode {
	switch {
	case pending:
		return listModePending
	case reviewed:
		return listModeReviewed
	case mentions:
		return listModeMentions
	default:
		return listModeMixed
	}
}

func sortPRs(prs []github.PR, mode listMode) {
	sort.SliceStable(prs, func(i, j int) bool {
		left := prPriority(prs[i], mode)
		right := prPriority(prs[j], mode)
		if left != right {
			return left < right
		}

		// Older PRs are more likely to be stale and need attention first.
		if !prs[i].CreatedAt.Equal(prs[j].CreatedAt) {
			return prs[i].CreatedAt.Before(prs[j].CreatedAt)
		}

		leftAuthor := authorPriority(prs[i].Author)
		rightAuthor := authorPriority(prs[j].Author)
		if leftAuthor != rightAuthor {
			return leftAuthor < rightAuthor
		}

		if prs[i].Repo != prs[j].Repo {
			return prs[i].Repo < prs[j].Repo
		}

		return prs[i].Number < prs[j].Number
	})
}

func prPriority(pr github.PR, mode listMode) int {
	draftPenalty := 0
	if pr.IsDraft {
		draftPenalty = 10
	}

	switch mode {
	case listModePending:
		return draftPenalty
	case listModeReviewed:
		if hasNewActivity(pr) {
			return draftPenalty
		}
		return draftPenalty + 1
	case listModeMentions:
		return draftPenalty
	default:
		switch {
		case pr.Status == "pending":
			return draftPenalty
		case pr.Status == "reviewed" && hasNewActivity(pr):
			return draftPenalty + 1
		case pr.Status == "mentioned":
			return draftPenalty + 2
		case pr.Status == "reviewed":
			return draftPenalty + 3
		default:
			return draftPenalty + 4
		}
	}
}

func hasNewActivity(pr github.PR) bool {
	return strings.TrimSpace(pr.Activity) != ""
}

func authorPriority(author string) int {
	if strings.Contains(strings.ToLower(author), "dependabot") {
		return 1
	}
	return 0
}
