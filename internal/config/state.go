package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"gopkg.in/yaml.v3"
)

func NewState(packages shared.Packages) error {
	state := shared.State{
		Timestamp: time.Now(),
		Packages:  packages,
	}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(&state)
	if err != nil {
		return err
	}
	return os.WriteFile(StateFile, b.Bytes(), 0755)
}

func ReadState() (state shared.State, err error) {
	err = CreateOrMigrateStateFile()
	if err != nil {
		return
	}

	yamlRaw, err := os.ReadFile(StateFile)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlRaw, &state)
	if err != nil {
		return
	}
	return
}

func CreateOrMigrateStateFile() error {
	err := os.MkdirAll(DataDir, os.ModePerm)
	if err != nil {
		return err
	}

	info, err := os.Stat(StateFile)
	if os.IsNotExist(err) {
		return createStateFile()
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", PackageFile)
	}
	return nil
}

func createStateFile() error {
	bytes, err := yaml.Marshal(shared.State{Timestamp: time.Now()})
	if err != nil {
		return err
	}
	return os.WriteFile(StateFile, bytes, 0755)
}
