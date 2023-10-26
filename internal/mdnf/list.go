package mdnf

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/mdnf/internal/dnf"
	"github.com/lucas-ingemar/mdnf/internal/shared"
)

func List(ctx context.Context, packages shared.Packages) (installedPkgs []string, missingPkgs []string, err error) {
	dnfList, err := dnf.ListInstalled(ctx)
	if err != nil {
		return
	}

	for _, pkg := range packages.Global.Packages {
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

	return
}
