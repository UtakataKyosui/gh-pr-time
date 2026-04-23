package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCommitEntryDate_flat(t *testing.T) {
	var e CommitEntry
	if err := json.Unmarshal([]byte(`{"committedDate":"2024-01-15T10:00:00Z"}`), &e); err != nil {
		t.Fatal(err)
	}
	if got := e.date(); got != "2024-01-15T10:00:00Z" {
		t.Errorf("got %q, want %q", got, "2024-01-15T10:00:00Z")
	}
}

func TestCommitEntryDate_wrapped(t *testing.T) {
	var e CommitEntry
	if err := json.Unmarshal([]byte(`{"commit":{"committedDate":"2024-01-15T10:00:00Z"}}`), &e); err != nil {
		t.Fatal(err)
	}
	if got := e.date(); got != "2024-01-15T10:00:00Z" {
		t.Errorf("got %q, want %q", got, "2024-01-15T10:00:00Z")
	}
}

func TestCommitEntryDate_flatTakesPrecedence(t *testing.T) {
	data := `{"committedDate":"2024-01-15T10:00:00Z","commit":{"committedDate":"2024-01-15T09:00:00Z"}}`
	var e CommitEntry
	if err := json.Unmarshal([]byte(data), &e); err != nil {
		t.Fatal(err)
	}
	if got := e.date(); got != "2024-01-15T10:00:00Z" {
		t.Errorf("flat field should take precedence, got %q", got)
	}
}

func TestCommitEntryDate_empty(t *testing.T) {
	var e CommitEntry
	if got := e.date(); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestPRUnmarshal_flat(t *testing.T) {
	data := `{
		"number": 42,
		"title": "test PR",
		"commits": [
			{"committedDate": "2024-01-15T09:00:00Z"},
			{"committedDate": "2024-01-15T11:00:00Z"}
		]
	}`
	var pr PR
	if err := json.Unmarshal([]byte(data), &pr); err != nil {
		t.Fatal(err)
	}
	if pr.Number != 42 {
		t.Errorf("number: got %d, want 42", pr.Number)
	}
	if len(pr.Commits) != 2 {
		t.Fatalf("commits: got %d, want 2", len(pr.Commits))
	}
	if got := pr.Commits[0].date(); got != "2024-01-15T09:00:00Z" {
		t.Errorf("first commit date: got %q", got)
	}
}

func TestPRUnmarshal_wrapped(t *testing.T) {
	data := `{
		"number": 1,
		"title": "old gh",
		"commits": [
			{"commit": {"committedDate": "2024-01-15T09:00:00Z"}},
			{"commit": {"committedDate": "2024-01-15T11:00:00Z"}}
		]
	}`
	var pr PR
	if err := json.Unmarshal([]byte(data), &pr); err != nil {
		t.Fatal(err)
	}
	if len(pr.Commits) != 2 {
		t.Fatalf("commits: got %d, want 2", len(pr.Commits))
	}
	if got := pr.Commits[0].date(); got != "2024-01-15T09:00:00Z" {
		t.Errorf("first commit date: got %q", got)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0h 0m"},
		{30 * time.Minute, "0h 30m"},
		{90 * time.Minute, "1h 30m"},
		{2*time.Hour + 5*time.Minute, "2h 5m"},
		{10*time.Hour + 59*time.Minute, "10h 59m"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestGroupByDay_singleDay(t *testing.T) {
	commits := []CommitEntry{
		{CommittedDate: "2024-01-15T09:00:00Z"},
		{CommittedDate: "2024-01-15T11:30:00Z"},
	}
	dayMap := groupByDay(commits)
	if len(dayMap) != 1 {
		t.Fatalf("got %d days, want 1", len(dayMap))
	}
	if len(dayMap["2024-01-15"]) != 2 {
		t.Errorf("2024-01-15: got %d commits, want 2", len(dayMap["2024-01-15"]))
	}
}

func TestGroupByDay_multipleDays(t *testing.T) {
	commits := []CommitEntry{
		{CommittedDate: "2024-01-15T09:00:00Z"},
		{CommittedDate: "2024-01-15T11:30:00Z"},
		{CommittedDate: "2024-01-16T10:00:00Z"},
	}
	dayMap := groupByDay(commits)
	if len(dayMap) != 2 {
		t.Fatalf("got %d days, want 2", len(dayMap))
	}
	if len(dayMap["2024-01-15"]) != 2 {
		t.Errorf("2024-01-15: got %d commits, want 2", len(dayMap["2024-01-15"]))
	}
	if len(dayMap["2024-01-16"]) != 1 {
		t.Errorf("2024-01-16: got %d commits, want 1", len(dayMap["2024-01-16"]))
	}
}

func TestGroupByDay_invalidDateSkipped(t *testing.T) {
	commits := []CommitEntry{
		{CommittedDate: "not-a-date"},
		{CommittedDate: "2024-01-15T09:00:00Z"},
	}
	dayMap := groupByDay(commits)
	if len(dayMap) != 1 {
		t.Fatalf("got %d days, want 1 (invalid date should be skipped)", len(dayMap))
	}
}

func TestGroupByDay_empty(t *testing.T) {
	dayMap := groupByDay(nil)
	if len(dayMap) != 0 {
		t.Errorf("got %d days, want 0", len(dayMap))
	}
}

func TestGroupByDay_timesAreParsedCorrectly(t *testing.T) {
	commits := []CommitEntry{
		{CommittedDate: "2024-01-15T09:00:00Z"},
		{CommittedDate: "2024-01-15T11:30:00Z"},
	}
	dayMap := groupByDay(commits)
	times := dayMap["2024-01-15"]
	want0, _ := time.Parse(time.RFC3339, "2024-01-15T09:00:00Z")
	want1, _ := time.Parse(time.RFC3339, "2024-01-15T11:30:00Z")
	if !times[0].Equal(want0) {
		t.Errorf("times[0] = %v, want %v", times[0], want0)
	}
	if !times[1].Equal(want1) {
		t.Errorf("times[1] = %v, want %v", times[1], want1)
	}
}
