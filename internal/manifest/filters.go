package manifest

import (
	"fmt"
	"os"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func Filter(pmManifest shared.PmManifest) (packages []string, dependencies []string, err error) {
	packages = append(packages, pmManifest.Global.Packages...)
	dependencies = append(dependencies, pmManifest.Global.Dependencies...)

	for _, c := range pmManifest.Conditional {
		switch c.Type {
		case shared.MConditionHost:
			p, d, err := filterHost(c)
			if err != nil {
				return nil, nil, err
			}
			packages = append(packages, p...)
			dependencies = append(dependencies, d...)
		case shared.MConditionGroup:
			p, d, err := filterGroup(c)
			if err != nil {
				return nil, nil, err
			}
			packages = append(packages, p...)
			dependencies = append(dependencies, d...)
		default:
			return nil, nil, fmt.Errorf("unknown condition type '%s'", c.Type)
		}
	}

	return
}

func filterHost(c shared.ManifestConditional) (packages []string, dependencies []string, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	if c.Value == hostname {
		return c.Packages, c.Dependencies, nil
	}
	return
}

func filterGroup(c shared.ManifestConditional) (packages []string, dependencies []string, err error) {
	if lo.Contains(config.Groups, c.Value) {
		return c.Packages, c.Dependencies, nil
	}
	return
}
