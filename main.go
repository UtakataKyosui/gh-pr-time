package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"
)

// CommitEntry accepts both JSON shapes emitted by different gh versions:
//   flat (newer):   {"committedDate": "..."}
//   wrapped (older): {"commit": {"committedDate": "..."}}
type CommitEntry struct {
	CommittedDate string `json:"committedDate"`
	Commit        struct {
		CommittedDate string `json:"committedDate"`
	} `json:"commit"`
}

func (e CommitEntry) date() string {
	if e.CommittedDate != "" {
		return e.CommittedDate
	}
	return e.Commit.CommittedDate
}

type PR struct {
	Number  int           `json:"number"`
	Title   string        `json:"title"`
	Commits []CommitEntry `json:"commits"`
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

	dayMap := groupByDay(pr.Commits)

	keys := make([]string, 0, len(dayMap))
	for k := range dayMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("#%d %s\n\n", pr.Number, pr.Title)

	var total time.Duration
	for _, day := range keys {
		times := dayMap[day]
		sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })

		first, last := times[0], times[len(times)-1]
		diff := last.Sub(first)
		total += diff

		fmt.Printf("%s  %s - %s  (%s)\n", day, first.Format("15:04"), last.Format("15:04"), formatDuration(diff))
	}

	fmt.Printf("\nTotal active time: %s\n", formatDuration(total))
}

func groupByDay(commits []CommitEntry) map[string][]time.Time {
	dayMap := map[string][]time.Time{}
	for _, c := range commits {
		t, err := time.Parse(time.RFC3339, c.date())
		if err != nil {
			continue
		}
		key := t.Format("2006-01-02")
		dayMap[key] = append(dayMap[key], t)
	}
	return dayMap
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}