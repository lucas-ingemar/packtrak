package config

import (
	"os"

	"github.com/lucas-ingemar/mdnf/internal/shared"

	"gopkg.in/yaml.v3"
)

func ReadPackagesConfig(filename string) (packages shared.Packages, err error) {
	yamlRaw, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlRaw, &packages)
	if err != nil {
		return
	}
	return
}
