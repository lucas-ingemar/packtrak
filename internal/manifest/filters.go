package manifest

import (
	"fmt"
	"os"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/samber/lo"
)

func MatchConditional(c Conditional) (match bool, err error) {
	switch c.Type {
	case MConditionHost:
		return filterHost(c)
	case MConditionGroup:
		return filterGroup(c)
	default:
		return false, fmt.Errorf("unknown condition type '%s'", c.Type)
	}
}

func Filter(pmManifest PmManifest) (packages []string, dependencies []string, err error) {
	packages = append(packages, pmManifest.Global.Packages...)
	dependencies = append(dependencies, pmManifest.Global.Dependencies...)

	for _, c := range pmManifest.Conditional {
		switch c.Type {
		case MConditionHost:
			match, err := filterHost(c)
			if err != nil {
				return nil, nil, err
			}
			if match {
				packages = append(packages, c.Packages...)
				dependencies = append(dependencies, c.Dependencies...)
			}
		case MConditionGroup:
			match, err := filterGroup(c)
			if err != nil {
				return nil, nil, err
			}
			if match {
				packages = append(packages, c.Packages...)
				dependencies = append(dependencies, c.Dependencies...)
			}
		default:
			return nil, nil, fmt.Errorf("unknown condition type '%s'", c.Type)
		}
	}

	return
}

func filterHost(c Conditional) (match bool, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	if c.Value == hostname {
		return true, nil
		// return c.Packages, c.Dependencies, nil
	}
	return
}

func filterGroup(c Conditional) (match bool, err error) {
	if lo.Contains(config.Groups, c.Value) {
		return true, nil
		// return c.Packages, c.Dependencies, nil
	}
	return
}
