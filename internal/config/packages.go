package config

// import (
// 	"bytes"
// 	"fmt"
// 	"os"

// 	"github.com/lucas-ingemar/packtrak/internal/shared"

// 	"gopkg.in/yaml.v3"
// )

// var (
// 	Packages shared.Packages
// )

// func SavePackages() error {
// 	var b bytes.Buffer
// 	yamlEncoder := yaml.NewEncoder(&b)
// 	yamlEncoder.SetIndent(2)
// 	err := yamlEncoder.Encode(&Packages)
// 	if err != nil {
// 		return err
// 	}
// 	return os.WriteFile(PackageFile, b.Bytes(), 0755)
// }

// func ReadPackagesConfig() (packages shared.Packages, err error) {
// 	err = CreateOrMigratePackageFile()
// 	if err != nil {
// 		return
// 	}

// 	yamlRaw, err := os.ReadFile(PackageFile)
// 	if err != nil {
// 		return
// 	}

// 	err = yaml.Unmarshal(yamlRaw, &packages)
// 	if err != nil {
// 		return
// 	}
// 	return
// }

// func CreateOrMigratePackageFile() error {
// 	err := os.MkdirAll(ConfigDir, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}

// 	info, err := os.Stat(PackageFile)
// 	if os.IsNotExist(err) {
// 		return createPackagesFile()
// 	}

// 	if info.IsDir() {
// 		return fmt.Errorf("%s is a directory", PackageFile)
// 	}

// 	return nil
// }

// func createPackagesFile() error {
// 	bytes, err := yaml.Marshal(shared.Packages{})
// 	if err != nil {
// 		return err
// 	}
// 	return os.WriteFile(PackageFile, bytes, 0755)
// }
