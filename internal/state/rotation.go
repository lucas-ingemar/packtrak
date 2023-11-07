package state

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/samber/lo"
)

func Rotate(rotations int) error {
	if rotations <= 0 {
		return nil
	}
	files, err := os.ReadDir(config.DataDir)
	if err != nil {
		return err
	}
	dbFiles := []string{}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".db") {
			dbFiles = append(dbFiles, filepath.Join(config.DataDir, f.Name()))
		}
	}

	if !lo.Contains(dbFiles, config.StateFile) {
		return fmt.Errorf("%s not found", config.StateFile)
	}
	dbFiles = lo.Filter(dbFiles, func(item string, index int) bool { return item != config.StateFile })

	if len(dbFiles) < rotations {
		return copyStatefile()
	}

	sort.Strings(dbFiles)
	nToRemove := len(dbFiles) - rotations + 1 // We want to remove one extra file to add place for the new one
	for _, f := range dbFiles[0:nToRemove] {
		if err := os.Remove(f); err != nil {
			return nil
		}
	}

	return copyStatefile()
}

func copyStatefile() error {
	timeStr := time.Now().Format("20060102T150405")
	stateRotFile := path.Join(config.DataDir, fmt.Sprintf("state.%s.db", timeStr))

	data, err := os.ReadFile(config.StateFile)
	if err != nil {
		return err
	}

	return os.WriteFile(stateRotFile, data, 0644)
}
