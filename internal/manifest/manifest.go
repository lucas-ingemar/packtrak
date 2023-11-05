package manifest

import (
	"bytes"
	"fmt"
	"os"

	"github.com/lucas-ingemar/packtrak/internal/shared"

	"gopkg.in/yaml.v3"
)

var (
	Manifest shared.Manifest
)

func SaveManifest(filename string) error {
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(&Manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b.Bytes(), 0755)
}

func ReadManifest(filename string) (manifest shared.Manifest, err error) {
	err = CreateOrMigrateManifestFile(filename)
	if err != nil {
		return
	}

	yamlRaw, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlRaw, &manifest)
	if err != nil {
		return
	}
	return
}

func CreateOrMigrateManifestFile(filename string) error {
	// err := os.MkdirAll(ConfigDir, os.ModePerm)
	// if err != nil {
	// 	return err
	// }

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return createManifestFile(filename)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}

	return nil
}

func createManifestFile(filename string) error {
	bytes, err := yaml.Marshal(shared.Manifest{})
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0755)
}
