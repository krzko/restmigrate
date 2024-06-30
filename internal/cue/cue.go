package cue

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/krzko/restmigrate/internal/migration"
)

func ParseMigration(filename string) ([]migration.Migration, error) {
	ctx := cuecontext.New()
	instances := load.Instances([]string{filename}, nil)
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found")
	}

	value := ctx.BuildInstance(instances[0])
	if value.Err() != nil {
		return nil, value.Err()
	}

	var migrations []migration.Migration
	err := value.LookupPath(cue.ParsePath("migrations")).Decode(&migrations)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}
