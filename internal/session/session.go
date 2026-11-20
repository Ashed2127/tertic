package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Ashed2127/tertic/internal/openrouter"
)

type SavedSession struct {
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	Provider  string               `json:"provider"`
	Model     string               `json:"model"`
	WorkDir   string               `json:"work_dir"`
	History   []openrouter.Message `json:"history"`
}

func GetSessionsDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}
	dir := filepath.Join(home, ".tertic", "sessions")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func SaveSession(s *SavedSession) error {
	if s == nil || s.ID == "" {
		return fmt.Errorf("invalid session")
	}
	s.UpdatedAt = time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = s.UpdatedAt
	}
	if s.Title == "" {
		// derive title from first user message
		for _, m := range s.History {
			if m.Role == "user" && m.Content != "" {
				s.Title = strings.TrimSpace(m.Content)
				if len(s.Title) > 40 {
					s.Title = s.Title[:37] + "..."
				}
				break
			}
		}
		if s.Title == "" {
			s.Title = "Session " + s.ID
		}
	}

	path := filepath.Join(GetSessionsDir(), s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadSession(id string) (*SavedSession, error) {
	if !strings.HasSuffix(id, ".json") {
		id += ".json"
	}
	path := filepath.Join(GetSessionsDir(), id)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s SavedSession
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func LoadLatestSession() (*SavedSession, error) {
	list, err := ListSessions()
	if err != nil || len(list) == 0 {
		return nil, fmt.Errorf("no existing sessions found")
	}
	return LoadSession(list[0].ID)
}

func ListSessions() ([]SavedSession, error) {
	dir := GetSessionsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var list []SavedSession
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s SavedSession
		if err := json.Unmarshal(data, &s); err == nil && s.ID != "" {
			list = append(list, s)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})

	return list, nil
}
