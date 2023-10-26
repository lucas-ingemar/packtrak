package shared

type Packages struct {
	Global struct {
		Packages []string `yaml:"packages"`
	} `yaml:"global"`
}
