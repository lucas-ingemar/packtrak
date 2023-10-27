package shared

import "time"

type Packages struct {
	Dnf Dnf `yaml:"dnf"`
}

type Dnf struct {
	Global DnfPackagesGlobal `yaml:"global"`
}

type DnfPackagesGlobal struct {
	Packages []string `yaml:"packages"`
}

type State struct {
	Timestamp time.Time `yaml:"timestamp"`
	Packages  Packages  `yaml:"packages"`
}
