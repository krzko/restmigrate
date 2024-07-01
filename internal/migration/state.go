package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/krzko/restmigrate/internal/logger"
	"github.com/krzko/restmigrate/internal/telemetry"
)

type AppliedMigration struct {
	Timestamp int64  `json:"timestamp"`
	Name      string `json:"name"`
}

type State struct {
	AppVersion        string             `json:"app_version"`
	AppliedMigrations []AppliedMigration `json:"applied_migrations"`
}

const stateFileName = "restmigrate.state"

func LoadState(ctx context.Context, path, appVersion string) (*State, error) {
	ctx, span := telemetry.StartSpan(ctx, "LoadState")
	defer span.End()

	stateFilePath := filepath.Join(path, stateFileName)
	logger.Debug("Loading state file", "path", stateFilePath)

	data, err := os.ReadFile(stateFilePath)
	if os.IsNotExist(err) {
		logger.Info("State file not found, creating new state", "path", stateFilePath)
		return &State{AppVersion: appVersion}, nil
	} else if err != nil {
		return nil, err
	}

	var state State
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}

	// Update the app version if it's different
	if state.AppVersion != appVersion {
		logger.Info("Updating app version in state file", "old", state.AppVersion, "new", appVersion)
		state.AppVersion = appVersion
		err = state.Save(path)
		if err != nil {
			return nil, fmt.Errorf("failed to update app version in state file: %w", err)
		}
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
