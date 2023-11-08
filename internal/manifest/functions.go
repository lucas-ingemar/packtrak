package manifest

import (
	"os"

	"github.com/lucas-ingemar/packtrak/internal/shared"
)

// FIXME: Somhow os.Hostname needs to be moved to an interface
func AddToHost(toAdd []string, pmName string, installDependency bool) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	mc, err := Manifest.Pm(pmName).GetOrAddConditional(shared.MConditionHost, hostname)
	if err != nil {
		return err
	}
	if installDependency {
		mc.AddDependencies(toAdd)
	} else {
		mc.AddPackages(toAdd)
	}
	return nil
}

func AddToGroup(toAdd []string, pmName string, group string, installDependency bool) error {
	mc, err := Manifest.Pm(pmName).GetOrAddConditional(shared.MConditionGroup, group)
	if err != nil {
		return err
	}
	if installDependency {
		mc.AddDependencies(toAdd)
	} else {
		mc.AddPackages(toAdd)
	}
	return nil
}
