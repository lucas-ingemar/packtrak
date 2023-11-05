package shared

import (
	"fmt"
	"time"
)

type Packages map[string]PmPackages

func (p Packages) Register(packageManagerName string) error {
	_, exists := p[packageManagerName]
	if exists {
		return fmt.Errorf("%s already exists", packageManagerName)
	}
	p[packageManagerName] = PmPackages{}
	return nil
}

type PmPackages struct {
	Global PackagesGlobal `yaml:"global"`
}

type PackagesGlobal struct {
	Dependencies []string `yaml:"dependencies"`
	Packages     []string `yaml:"packages"`
}

type State struct {
	Timestamp time.Time `yaml:"timestamp"`
	Packages  Packages  `yaml:"packages"`
}

type Package struct {
	Name          string
	FullName      string
	Version       string
	LatestVersion string
	RepoUrl       string
}

type PackageStatus struct {
	Synced  []Package
	Updated []Package
	Missing []Package
	Removed []Package
}

type Dependency struct {
	Name     string
	FullName string
	// Version       string
	// LatestVersion string
	// RepoUrl       string
}

type DependenciesStatus struct {
	Synced  []Dependency
	Updated []Dependency
	Missing []Dependency
	Removed []Dependency
}

type CommandName string

const (
	CommandInstall CommandName = "install"
	CommandRemove  CommandName = "remove"
	CommandList    CommandName = "list"
	CommandSync    CommandName = "sync"
)
