package shared

import (
	"fmt"
	"os"
)

func IsSudo() bool {
	if os.Getenv("SUDO_UID") != "" && os.Getenv("SUDO_GID") != "" && os.Getenv("SUDO_USER") != "" {
		return true
	}
	return false
}

func GetPackage(name string, packages []Package) (Package, error) {
	for _, pkg := range packages {
		if pkg.Name == name {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("package %s not found", name)
}
