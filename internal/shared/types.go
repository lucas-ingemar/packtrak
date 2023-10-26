package shared

type Packages struct {
	Global PackagesGlobal `yaml:"global"`
}

type PackagesGlobal struct {
	Packages []string `yaml:"packages"`
}
