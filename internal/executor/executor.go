package executor

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/krzko/restmigrate/internal/cue"
	"github.com/krzko/restmigrate/internal/logger"
	"github.com/krzko/restmigrate/internal/migration"
	"github.com/krzko/restmigrate/pkg/rest"
	"github.com/urfave/cli/v2"
)

func ExecuteUp(c *cli.Context) error {
	logger.Debug("Starting ExecuteUp")
	path := c.String("path")

	state, err := migration.LoadState(path)
	if err != nil {
		logger.Error("Failed to load state", "error", err)
		return fmt.Errorf("failed to load state: %w", err)
	}

	migrations, err := loadMigrations(path)
	if err != nil {
		logger.Error("Failed to load migrations", "error", err)
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	client := rest.NewClient(c.String("base-url"), c.String("api-key"))

	for _, m := range migrations {
		if !containsMigration(state.AppliedMigrations, m.Timestamp) {
			logger.Info("Applying migration", "name", m.Name)
			err := applyMigration(client, m.Up)
			if err != nil {
				logger.Error("Failed to apply migration", "name", m.Name, "error", err)
				return fmt.Errorf("failed to apply migration %s: %w", m.Name, err)
			}
			state.AddMigration(m.Timestamp, m.Name)
			err = state.Save(path)
			if err != nil {
				logger.Error("Failed to save state", "error", err)
				return fmt.Errorf("failed to save state: %w", err)
			}
			logger.Info("Successfully applied migration", "name", m.Name)
		} else {
			logger.Debug("Skipping already applied migration", "name", m.Name)
		}
	}

	logger.Info("All migrations have been applied")
	return nil
}

func ExecuteDown(c *cli.Context) error {
	logger.Debug("Starting ExecuteDown")
	path := c.String("path")

	state, err := migration.LoadState(path)
	if err != nil {
		logger.Error("Failed to load state", "error", err)
		return fmt.Errorf("failed to load state: %w", err)
	}

	if len(state.AppliedMigrations) == 0 {
		logger.Info("No migrations to revert")
		return nil
	}

	client := rest.NewClient(c.String("base-url"), c.String("api-key"))

	if c.Bool("all") {
		logger.Info("Reverting all migrations")
		return revertAllMigrations(state, client, path)
	}

	logger.Info("Reverting last migration")
	return revertLastMigration(state, client, path)
}

func revertAllMigrations(state *migration.State, client *rest.Client, path string) error {
	logger.Debug("Starting revertAllMigrations")

	for i := len(state.AppliedMigrations) - 1; i >= 0; i-- {
		appliedMigration := state.AppliedMigrations[i]
		m, err := loadMigration(path, appliedMigration.Timestamp)
		if err != nil {
			logger.Error("Failed to load migration", "timestamp", appliedMigration.Timestamp, "error", err)
			return fmt.Errorf("failed to load migration: %w", err)
		}

		logger.Info("Reverting migration", "name", m.Name)
		err = applyMigration(client, m.Down)
		if err != nil {
			logger.Error("Failed to revert migration", "name", m.Name, "error", err)
			return fmt.Errorf("failed to revert migration %s: %w", m.Name, err)
		}

		state.RemoveLastMigration()
		err = state.Save(path)
		if err != nil {
			logger.Error("Failed to save state", "error", err)
			return fmt.Errorf("failed to save state: %w", err)
		}

		logger.Info("Successfully reverted migration", "name", m.Name)
	}

	logger.Info("All migrations have been reverted")
	return nil
}

func revertLastMigration(state *migration.State, client *rest.Client, path string) error {
	logger.Debug("Starting revertLastMigration")

	lastMigration := state.AppliedMigrations[len(state.AppliedMigrations)-1]
	m, err := loadMigration(path, lastMigration.Timestamp)
	if err != nil {
		logger.Error("Failed to load migration", "timestamp", lastMigration.Timestamp, "error", err)
		return fmt.Errorf("failed to load migration: %w", err)
	}

	logger.Info("Reverting migration", "name", m.Name)
	err = applyMigration(client, m.Down)
	if err != nil {
		logger.Error("Failed to revert migration", "name", m.Name, "error", err)
		return fmt.Errorf("failed to revert migration %s: %w", m.Name, err)
	}

	state.RemoveLastMigration()
	err = state.Save(path)
	if err != nil {
		logger.Error("Failed to save state", "error", err)
		return fmt.Errorf("failed to save state: %w", err)
	}

	logger.Info("Successfully reverted last migration", "name", m.Name)
	return nil
}

func loadMigrations(path string) ([]migration.Migration, error) {
	logger.Debug("Loading migrations", "path", path)

	files, err := filepath.Glob(filepath.Join(path, "migrations/*.cue"))
	if err != nil {
		return nil, err
	}

	var allMigrations []migration.Migration
	for _, file := range files {
		logger.Debug("Parsing migration file", "file", file)
		migrations, err := cue.ParseMigration(file)
		if err != nil {
			logger.Error("Failed to parse migration file", "file", file, "error", err)
			return nil, fmt.Errorf("failed to parse migration %s: %w", file, err)
		}
		allMigrations = append(allMigrations, migrations...)
	}

	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].Timestamp < allMigrations[j].Timestamp
	})

	logger.Debug("Loaded migrations", "count", len(allMigrations))
	return allMigrations, nil
}

func loadMigration(path string, timestamp int64) (*migration.Migration, error) {
	logger.Debug("Loading migration", "timestamp", timestamp)

	migrations, err := loadMigrations(path)
	if err != nil {
		return nil, err
	}

	for i, m := range migrations {
		if m.Timestamp == timestamp {
			return &migrations[i], nil
		}
	}

	return nil, fmt.Errorf("migration not found for timestamp %d", timestamp)
}

func applyMigration(client *rest.Client, actions map[string]interface{}) error {
	logger.Debug("Applying migration actions")

	for endpoint, action := range actions {
		actionMap, ok := action.(map[string]interface{})
		if !ok {
			logger.Error("Invalid action format", "endpoint", endpoint)
			return fmt.Errorf("invalid action format for endpoint %s", endpoint)
		}

		method, ok := actionMap["method"].(string)
		if !ok {
			logger.Error("Missing or invalid method", "endpoint", endpoint)
			return fmt.Errorf("missing or invalid method for endpoint %s", endpoint)
		}

		var body interface{}
		if bodyData, exists := actionMap["body"]; exists {
			body = bodyData
		}

		logger.Debug("Sending request", "method", method, "endpoint", endpoint)
		err := client.SendRequest(method, endpoint, body)
		if err != nil {
			logger.Error("Failed to apply action", "endpoint", endpoint, "error", err)
			return fmt.Errorf("failed to apply action for endpoint %s: %w", endpoint, err)
		}
		logger.Debug("Successfully applied action", "endpoint", endpoint)
	}
	return nil
}

func containsMigration(appliedMigrations []migration.AppliedMigration, timestamp int64) bool {
	for _, m := range appliedMigrations {
		if m.Timestamp == timestamp {
			return true
		}
	}
	return false
}
