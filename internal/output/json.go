package output

import (
	"encoding/json"
	"fmt"

	"github.com/amiraminb/gh-plantir/internal/github"
)

func JSON(prs []github.PR) {
	data, err := json.MarshalIndent(prs, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
