package shared

import (
	"fmt"

	"github.com/samber/lo"
)

type (
	CommandName             string
	ManifestConditionalType string
	// Manifest                map[string]PmManifest
)

const (
	CommandInstall CommandName = "install"
	CommandRemove  CommandName = "remove"
	CommandList    CommandName = "list"
	CommandSync    CommandName = "sync"

	MConditionHost  ManifestConditionalType = "host"
	MConditionGroup ManifestConditionalType = "group"
)

// type Packages map[string]PmPackages

// func (p Packages) Register(packageManagerName string) error {
// 	_, exists := p[packageManagerName]
// 	if exists {
// 		return fmt.Errorf("%s already exists", packageManagerName)
// 	}
// 	p[packageManagerName] = PmPackages{}
// 	return nil
// }

type Manifest struct {
	Dnf PmManifest `yaml:"dnf"`
	Go  PmManifest `yaml:"go"`
}

func (m *Manifest) Pm(name string) *PmManifest {
	switch name {
	case "dnf":
		return &m.Dnf
	case "go":
		return &m.Go
	default:
		panic(fmt.Sprintf("%s is not a registered package manager", name))
	}
}

type PmManifest struct {
	Global      ManifestGlobal        `yaml:"global"`
	Conditional []ManifestConditional `yaml:"conditional"`
}

type ManifestGlobal struct {
	Dependencies []string `yaml:"dependencies"`
	Packages     []string `yaml:"packages"`
}

func (m *ManifestGlobal) AddPackages(packages []string) {
	m.Packages = append(m.Packages, packages...)
}

func (m *ManifestGlobal) RemovePackages(packages []string) {
	m.Packages = lo.Filter(m.Packages, func(item string, index int) bool {
		return !lo.Contains(packages, item)
	})
}

func (m *ManifestGlobal) AddDependencies(deps []string) {
	m.Dependencies = append(m.Dependencies, deps...)
}

func (m *ManifestGlobal) RemoveDependencies(deps []string) {
	m.Dependencies = lo.Filter(m.Dependencies, func(item string, index int) bool {
		return !lo.Contains(deps, item)
	})
}

type ManifestConditional struct {
	Type         ManifestConditionalType `yaml:"type"`
	Value        string                  `yaml:"value"`
	Dependencies []string                `yaml:"dependencies"`
	Packages     []string                `yaml:"packages"`
}

// type State struct {
// 	Timestamp time.Time `yaml:"timestamp"`
// 	Packages  Packages  `yaml:"packages"`
// }

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
