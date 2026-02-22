package generator

import "testing"

func TestParseCandidatesTruncatesToFive(t *testing.T) {
	raw := `{
  "candidates": [
    {"command":"c1","title":"t1","description":"d1","args":[]},
    {"command":"c2","title":"t2","description":"d2","args":[]},
    {"command":"c3","title":"t3","description":"d3","args":[]},
    {"command":"c4","title":"t4","description":"d4","args":[]},
    {"command":"c5","title":"t5","description":"d5","args":[]},
    {"command":"c6","title":"t6","description":"d6","args":[]}
  ]
}`

	got, err := parseCandidates(raw)
	if err != nil {
		t.Fatalf("parseCandidates returned error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 candidates, got %d", len(got))
	}
	if got[4].Command != "c5" {
		t.Fatalf("expected candidate 5 to be c5, got %q", got[4].Command)
	}
}

func TestParseCandidatesSupportsMarkdownFences(t *testing.T) {
	raw := "```json\n{\"candidates\":[{\"command\":\"ls -la\",\"title\":\"list\",\"description\":\"show files\",\"args\":[{\"name\":\"-la\",\"example\":\"ls -la\",\"meaning\":\"show all\"}]}]}\n```"
	got, err := parseCandidates(raw)
	if err != nil {
		t.Fatalf("parseCandidates returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	if got[0].Args[0].Name != "-la" {
		t.Fatalf("expected arg name -la, got %q", got[0].Args[0].Name)
	}
}

func TestParseCandidatesRejectsInvalidRows(t *testing.T) {
	raw := `{
  "candidates": [
    {"command":"","title":"bad","description":"missing command","args":[]},
    {"command":"echo ok","title":"ok","description":"valid","args":[]}
  ]
}`

	got, err := parseCandidates(raw)
	if err != nil {
		t.Fatalf("parseCandidates returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 valid candidate, got %d", len(got))
	}
	if got[0].Command != "echo ok" {
		t.Fatalf("expected echo ok, got %q", got[0].Command)
	}
}
