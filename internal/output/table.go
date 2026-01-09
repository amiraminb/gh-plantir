package output

import (
	"os"
	"strconv"
	"time"

	"github.com/amiraminb/plantir/internal/github"
	"github.com/olekukonko/tablewriter"
)

func age(t time.Time) string {
	d := time.Since(t)

	if d.Hours() >= 24 {
		days := int(d.Hours() / 24)
		return strconv.Itoa(days) + "d"
	}
	if d.Hours() >= 1 {
		return strconv.Itoa(int(d.Hours())) + "h"
	}
	return strconv.Itoa(int(d.Minutes())) + "m"
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
			pr.Type(),
		}

		if hasStatus {
			status := pr.Status
			if status == "" {
				status = "-"
			}
			row = append(row, status)
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
