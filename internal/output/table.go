package output

import (
	"os"
	"strconv"
	"time"

	"github.com/amiraminb/plantir/internal/github"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var (
	// Type colors
	dependabotColor = color.New(color.FgBlue).SprintFunc()
	featureColor    = color.New(color.FgMagenta).SprintFunc()

	// Age colors
	freshColor = color.New(color.FgHiCyan).SprintFunc() // < 1 day
	staleColor = color.New(color.FgYellow).SprintFunc() // 1-7 days
	oldColor   = color.New(color.FgRed).SprintFunc()    // > 7 days

	// Status colors
	reviewedColor = color.New(color.FgHiCyan).SprintFunc()
	pendingColor  = color.New(color.FgYellow).SprintFunc()
)

func age(t time.Time) string {
	d := time.Since(t)
	days := int(d.Hours() / 24)

	var ageStr string
	if d.Hours() >= 24 {
		ageStr = strconv.Itoa(days) + "d"
	} else if d.Hours() >= 1 {
		ageStr = strconv.Itoa(int(d.Hours())) + "h"
	} else {
		ageStr = strconv.Itoa(int(d.Minutes())) + "m"
	}

	// Color based on age
	switch {
	case days < 1:
		return freshColor(ageStr)
	case days <= 7:
		return staleColor(ageStr)
	default:
		return oldColor(ageStr)
	}
}

func coloredType(prType string) string {
	switch prType {
	case "dependabot":
		return dependabotColor(prType)
	default:
		return featureColor(prType)
	}
}

func coloredStatus(status string) string {
	switch status {
	case "reviewed":
		return reviewedColor(status)
	case "pending":
		return pendingColor(status)
	default:
		return status
	}
}

func Table(prs []github.PR) {
	hasActivity := false
	hasStatus := false
	for _, pr := range prs {
		if pr.Activity != "" {
			hasActivity = true
		}
		if pr.Status != "" {
			hasStatus = true
		}
	}

	table := tablewriter.NewTable(os.Stdout)

	// Build header based on available fields
	if hasStatus && hasActivity {
		table.Header("Repo", "PR#", "Title", "Author", "Age", "Type", "Status", "Activity")
	} else if hasStatus {
		table.Header("Repo", "PR#", "Title", "Author", "Age", "Type", "Status")
	} else if hasActivity {
		table.Header("Repo", "PR#", "Title", "Author", "Age", "Type", "Activity")
	} else {
		table.Header("Repo", "PR#", "Title", "Author", "Age", "Type")
	}

	for _, pr := range prs {
		title := pr.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}

		row := []string{
			pr.Repo,
			"#" + strconv.Itoa(pr.Number),
			title,
			pr.Author,
			age(pr.CreatedAt),
			coloredType(pr.Type()),
		}

		if hasStatus {
			status := pr.Status
			if status == "" {
				status = "-"
			}
			row = append(row, coloredStatus(status))
		}

		if hasActivity {
			activity := pr.Activity
			if activity == "" {
				activity = "-"
			}
			row = append(row, activity)
		}

		table.Append(row)
	}

	table.Render()
}
