package migration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/krzko/restmigrate/internal/logger"
)

type AppliedMigration struct {
	Timestamp int64  `json:"timestamp"`
	Name      string `json:"name"`
}

type State struct {
	AppliedMigrations []AppliedMigration `json:"applied_migrations"`
}

const stateFileName = "restmigrate.state.json"

func LoadState(path string) (*State, error) {
	stateFilePath := filepath.Join(path, stateFileName)
	logger.Debug("Loading state file", "path", stateFilePath)

	data, err := os.ReadFile(stateFilePath)
	if os.IsNotExist(err) {
		logger.Info("State file not found, creating new state", "path", stateFilePath)
		return &State{}, nil
	} else if err != nil {
		return nil, err
	}

	var state State
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (s *State) Save(path string) error {
	stateFilePath := filepath.Join(path, stateFileName)
	logger.Debug("Saving state file", "path", stateFilePath)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFilePath, data, 0644)
}

func (s *State) AddMigration(timestamp int64, name string) {
	s.AppliedMigrations = append(s.AppliedMigrations, AppliedMigration{
		Timestamp: timestamp,
		Name:      name,
	})
	sort.Slice(s.AppliedMigrations, func(i, j int) bool {
		return s.AppliedMigrations[i].Timestamp < s.AppliedMigrations[j].Timestamp
	})
}

func (s *State) RemoveLastMigration() {
	if len(s.AppliedMigrations) > 0 {
		s.AppliedMigrations = s.AppliedMigrations[:len(s.AppliedMigrations)-1]
	}
}
