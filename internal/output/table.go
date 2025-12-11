package output

import (
	"os"
	"strconv"
	"strings"
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
	table := tablewriter.NewTable(os.Stdout)
	table.Header("Repo", "PR#", "Title", "Author", "Age", "Type", "Labels")

	for _, pr := range prs {
		title := pr.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		labels := strings.Join(pr.Labels, ", ")
		if len(labels) > 25 {
			labels = labels[:22] + "..."
		}

		table.Append([]string{
			pr.Repo,
			"#" + strconv.Itoa(pr.Number),
			title,
			pr.Author,
			age(pr.CreatedAt),
			pr.Type(),
			labels,
		})
	}

	table.Render()
}
