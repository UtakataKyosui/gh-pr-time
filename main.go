package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"
)

type Commit struct {
	CommittedDate string `json:"committedDate"`
}

type CommitNode struct {
	Commit Commit `json:"commit"`
}

type PR struct {
	Number  int          `json:"number"`
	Title   string       `json:"title"`
	Commits []CommitNode `json:"commits"`
}

func main() {
	prArg := ""
	if len(os.Args) > 1 {
		prArg = os.Args[1]
	}

	args := []string{"pr", "view", "--json", "number,title,commits"}
	if prArg != "" {
		args = append(args, prArg)
	}

	out, err := exec.Command("gh", args...).Output()
	if err != nil {
		fmt.Println("failed:", err)
		os.Exit(1)
	}

	var pr PR
	if err := json.Unmarshal(out, &pr); err != nil {
		fmt.Println("json parse error:", err)
		os.Exit(1)
	}

	dayMap := map[string][]time.Time{}

	for _, c := range pr.Commits {
		t, err := time.Parse(time.RFC3339, c.Commit.CommittedDate)
		if err != nil {
			continue
		}

		dayKey := t.Format("2006-01-02")
		dayMap[dayKey] = append(dayMap[dayKey], t)
	}

	var total time.Duration

	keys := make([]string, 0, len(dayMap))
	for k := range dayMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("#%d %s\n\n", pr.Number, pr.Title)

	for _, day := range keys {
		times := dayMap[day]
		sort.Slice(times, func(i, j int) bool {
			return times[i].Before(times[j])
		})

		first := times[0]
		last := times[len(times)-1]
		diff := last.Sub(first)
		total += diff

		fmt.Printf("%s  %s - %s  (%s)\n",
			day,
			first.Format("15:04"),
			last.Format("15:04"),
			formatDuration(diff),
		)
	}

	fmt.Printf("\nTotal active time: %s\n", formatDuration(total))
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}