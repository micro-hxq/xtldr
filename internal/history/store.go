package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultLimit      = 200
	historyFileEnvKey = "XTLDR_HISTORY_FILE"
)

type Session struct {
	Request    string    `json:"request"`
	Command    string    `json:"command,omitempty"`
	WorkingDir string    `json:"working_dir,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Store struct {
	path  string
	limit int
}

func NewDefaultStore() (*Store, error) {
	path := strings.TrimSpace(os.Getenv(historyFileEnvKey))
	if path == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(configDir, "xtldr", "history.jsonl")
	}
	return &Store{path: path, limit: defaultLimit}, nil
}

func (s *Store) Append(request, command, workingDir string) error {
	request = strings.TrimSpace(request)
	command = strings.TrimSpace(command)
	if request == "" || command == "" {
		return nil
	}
	sessions, err := s.List()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	entry := Session{
		Request:    request,
		Command:    command,
		WorkingDir: strings.TrimSpace(workingDir),
		CreatedAt:  time.Now().UTC(),
	}
	filtered := make([]Session, 0, len(sessions))
	for _, existing := range sessions {
		if existing.Command == entry.Command {
			continue
		}
		filtered = append(filtered, existing)
	}
	sessions = append([]Session{entry}, filtered...)
	if len(sessions) > s.limit {
		sessions = sessions[:s.limit]
	}
	return s.writeAll(sessions)
}

func (s *Store) List() ([]Session, error) {
	file, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var sessions []Session
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry Session
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entry.Request = strings.TrimSpace(entry.Request)
		entry.Command = strings.TrimSpace(entry.Command)
		if entry.Request == "" {
			continue
		}
		sessions = append(sessions, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func Search(sessions []Session, query string) []Session {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return sessions
	}
	filtered := make([]Session, 0, len(sessions))
	for _, session := range sessions {
		if strings.Contains(strings.ToLower(session.Request), query) || strings.Contains(strings.ToLower(session.WorkingDir), query) {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

func (s *Store) writeAll(sessions []Session) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, session := range sessions {
		data, err := json.Marshal(session)
		if err != nil {
			continue
		}
		if _, err := writer.WriteString(string(data) + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
