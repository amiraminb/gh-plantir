package cmd

import (
	"slices"
	"testing"
	"time"

	"github.com/amiraminb/gh-plantir/internal/github"
)

func TestSortPRsMixedPrioritizesInboxWork(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	prs := []github.PR{
		{Number: 1, Repo: "api", Author: "alice", CreatedAt: now.Add(-48 * time.Hour), Status: "pending"},
		{Number: 2, Repo: "api", Author: "bob", CreatedAt: now.Add(-12 * time.Hour), Status: "reviewed", Activity: "2 commits"},
		{Number: 3, Repo: "api", Author: "carol", CreatedAt: now.Add(-72 * time.Hour), Status: "mentioned"},
		{Number: 4, Repo: "api", Author: "dependabot", CreatedAt: now.Add(-96 * time.Hour), Status: "reviewed"},
		{Number: 5, Repo: "api", Author: "dave", CreatedAt: now.Add(-120 * time.Hour), Status: "pending", IsDraft: true},
		{Number: 6, Repo: "api", Author: "erin", CreatedAt: now.Add(-168 * time.Hour), Status: "pending"},
	}

	sortPRs(prs, listModeMixed)

	got := prNumbers(prs)
	want := []int{6, 1, 2, 3, 4, 5}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortPRsReviewedPrioritizesNewActivityAndDraftsLast(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	prs := []github.PR{
		{Number: 1, Repo: "api", Author: "alice", CreatedAt: now.Add(-24 * time.Hour)},
		{Number: 2, Repo: "api", Author: "bob", CreatedAt: now.Add(-6 * time.Hour), Activity: "1 comment"},
		{Number: 3, Repo: "api", Author: "carol", CreatedAt: now.Add(-72 * time.Hour), IsDraft: true, Activity: "3 commits"},
		{Number: 4, Repo: "api", Author: "dave", CreatedAt: now.Add(-48 * time.Hour)},
	}

	sortPRs(prs, listModeReviewed)

	got := prNumbers(prs)
	want := []int{2, 4, 1, 3}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func prNumbers(prs []github.PR) []int {
	numbers := make([]int, len(prs))
	for i, pr := range prs {
		numbers[i] = pr.Number
	}
	return numbers
}
