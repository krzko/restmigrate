package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/krzko/restmigrate/internal/logger"
	"github.com/urfave/cli/v2"
)

type Migration struct {
	Timestamp int64                  `json:"timestamp"`
	Name      string                 `json:"name"`
	Up        map[string]interface{} `json:"up"`
	Down      map[string]interface{} `json:"down"`
}

func CreateMigration(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("migration name is required")
	}

	name := c.Args().First()
	now := time.Now()
	timestamp := now.Unix()
	dateString := now.Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.cue", dateString, name)

	content := fmt.Sprintf(`migrations: [
    {
        timestamp: %d
        name:      "%s"
        up: {
            // Define your up migration here
        }
        down: {
            // Define your down migration here
        }
    }
]
`, timestamp, name)

	path := c.String("path")
	filePath := filepath.Join(path, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	logger.Info("Created migration", "file", filePath)
	return nil
}
