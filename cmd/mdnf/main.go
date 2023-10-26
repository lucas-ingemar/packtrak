package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/lucas-ingemar/mdnf/internal/cmd"
	"github.com/lucas-ingemar/mdnf/internal/config"
	"github.com/lucas-ingemar/mdnf/internal/dnf"
)

func main() {
	cmd.Execute()
}

func main1() {
	packages, err := config.ReadPackagesConfig("packages_test.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println("Listing DNF packages...")
	dnfList, err := dnf.ListInstalled()
	if err != nil {
		panic(err)
	}

	installedPkgs := []string{}
	missingPkgs := []string{}
	for _, pkg := range packages.Global.Packages {
		pkgFound := false
		for _, dnfPkg := range dnfList {
			if strings.HasPrefix(dnfPkg, fmt.Sprintf("%s.", pkg)) {
				installedPkgs = append(installedPkgs, pkg)
				color.Green(" %s", pkg)
				pkgFound = true
				break
			}
		}
		if !pkgFound {
			missingPkgs = append(missingPkgs, pkg)
			color.Red(" %s", pkg)
		}
	}

	fmt.Println("")
	if len(missingPkgs) > 0 {
		color.Red("%d package(s) missing", len(missingPkgs))
	} else {
		color.Green("All packages installed")
	}

}
