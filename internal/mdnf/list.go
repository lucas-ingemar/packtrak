package mdnf

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/mdnf/internal/dnf"
	"github.com/lucas-ingemar/mdnf/internal/shared"
	"github.com/samber/lo"
)

func List(ctx context.Context, packages shared.Packages, state shared.State) (installedPkgs []string, missingPkgs []string, removedPkgs []string, err error) {
	dnfList, err := dnf.ListInstalled(ctx)
	if err != nil {
		return
	}

	for _, pkg := range packages.Dnf.Global.Packages {
		pkgFound := false
		for _, dnfPkg := range dnfList {
			if strings.HasPrefix(dnfPkg, fmt.Sprintf("%s.", pkg)) {
				installedPkgs = append(installedPkgs, pkg)
				pkgFound = true
				break
			}
		}
		if !pkgFound {
			missingPkgs = append(missingPkgs, pkg)
		}
	}

	for _, pkg := range state.Packages.Dnf.Global.Packages {
		for _, dnfPkg := range dnfList {
			if strings.HasPrefix(dnfPkg, fmt.Sprintf("%s.", pkg)) {
				if !lo.Contains(packages.Dnf.Global.Packages, pkg) {
					removedPkgs = append(removedPkgs, pkg)
				}
				break
			}
		}
	}

	return
}
