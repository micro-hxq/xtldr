package generator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	copilot "github.com/github/copilot-sdk/go"

	"xtldr/internal/model"
)

const maxCandidates = 5

type Copilot struct{}

func NewCopilot() *Copilot {
	return &Copilot{}
}

func (g *Copilot) Generate(ctx context.Context, request, workingDir string) ([]model.Candidate, error) {
	request = strings.TrimSpace(request)
	if request == "" {
		return nil, errors.New("request is required")
	}

	client := copilot.NewClient(&copilot.ClientOptions{Cwd: workingDir})
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("start copilot client: %w", err)
	}
	defer client.Stop()

	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		WorkingDirectory: workingDir,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Destroy()

	response, err := session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: buildPrompt(request),
	})
	if err != nil {
		return nil, fmt.Errorf("send prompt: %w", err)
	}
	if response == nil || response.Data.Content == nil {
		return nil, errors.New("copilot returned empty response")
	}

	candidates, err := parseCandidates(*response.Data.Content)
	if err != nil {
		return nil, fmt.Errorf("parse candidates: %w", err)
	}
	return candidates, nil
}

func buildPrompt(request string) string {
	return fmt.Sprintf(`You generate shell command candidates.
User request: %s

Return JSON only (no markdown fences, no extra text) using this schema:
{
  "candidates": [
    {
      "command": "string",
      "title": "string",
      "description": "string",
      "args": [
        {"name": "string", "example": "string", "meaning": "string"}
      ]
    }
  ]
}

Rules:
- Return 1 to 5 candidates only.
- Every command must be directly executable in a shell.
- Prefer safe and non-destructive commands.
- Include a clear command explanation and argument-level meanings.
`, request)
}

type candidateResponse struct {
	Candidates []model.Candidate `json:"candidates"`
}

func parseCandidates(raw string) ([]model.Candidate, error) {
	cleaned := extractJSONObject(raw)
	if cleaned == "" {
		return nil, errors.New("no JSON object in response")
	}

	var parsed candidateResponse
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, err
	}

	normalized := make([]model.Candidate, 0, maxCandidates)
	for _, candidate := range parsed.Candidates {
		c := normalizeCandidate(candidate)
		if c.Command == "" || c.Description == "" {
			continue
		}
		normalized = append(normalized, c)
		if len(normalized) == maxCandidates {
			break
		}
	}

	if len(normalized) == 0 {
		return nil, errors.New("no valid candidates")
	}
	return normalized, nil
}

func normalizeCandidate(candidate model.Candidate) model.Candidate {
	candidate.Command = strings.Trim(strings.TrimSpace(candidate.Command), "`")
	candidate.Title = strings.TrimSpace(candidate.Title)
	candidate.Description = strings.TrimSpace(candidate.Description)
	if candidate.Title == "" {
		candidate.Title = candidate.Command
	}

	args := make([]model.Argument, 0, len(candidate.Args))
	for _, arg := range candidate.Args {
		arg.Name = strings.TrimSpace(arg.Name)
		arg.Example = strings.TrimSpace(arg.Example)
		arg.Meaning = strings.TrimSpace(arg.Meaning)
		if arg.Name == "" && arg.Meaning == "" && arg.Example == "" {
			continue
		}
		if arg.Name == "" {
			arg.Name = "(positional)"
		}
		args = append(args, arg)
	}
	candidate.Args = args
	return candidate
}

func extractJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
			lines = lines[:len(lines)-1]
		}
		trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end <= start {
		return ""
	}
	return trimmed[start : end+1]
}
