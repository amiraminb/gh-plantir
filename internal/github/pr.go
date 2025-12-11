package github

import "time"

type PR struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Author    string    `json:"author"`
	Repo      string    `json:"repo"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	IsDraft   bool      `json:"isDraft"`
	Labels    []string  `json:"labels"`
}

func (p PR) Type() string {
	switch p.Author {
	case "dependabot":
		return "dependabot"
	default:
		return "feature"
	}
}
