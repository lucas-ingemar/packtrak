package app

import (
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

func (a App) PrintPackageList(depStatus map[managers.ManagerName]shared.DependenciesStatus, pkgStatus map[managers.ManagerName]shared.PackageStatus) error {
	noSynced, noUpdated, noMissing, noRemoved := 0, 0, 0, 0

	fmt.Println("\nDependencies:")
	for _, mName := range a.Managers.ListManagers() {
		m, err := a.Managers.GetManager(mName)
		if err != nil {
			return err
		}
		for _, dep := range depStatus[m.Name()].Synced {
			shared.PtermInstalled.Printfln("%s %s", m.Icon(), dep.Name)
			noSynced++
		}

		for _, dep := range depStatus[m.Name()].Missing {
			shared.PtermMissing.Printfln("%s %s", m.Icon(), dep.Name)
			noMissing++
		}

		for _, dep := range depStatus[m.Name()].Removed {
			shared.PtermRemoved.Printfln("%s %s", m.Icon(), dep.Name)
			noRemoved++
		}
	}

	fmt.Println("\nPackages:")
	for _, mName := range a.ListManagers() {
		m, err := a.Managers.GetManager(mName)
		if err != nil {
			return err
		}
		for _, pkg := range pkgStatus[m.Name()].Synced {
			shared.PtermInstalled.Printfln("%s %s", m.Icon(), pkg.Name)
			noSynced++
		}

		for _, pkg := range pkgStatus[m.Name()].Updated {
			shared.PtermUpdated.Printfln("%s %s %s -> %s", m.Icon(), pkg.Name, pkg.Version, pkg.LatestVersion)
			noUpdated++
		}

		for _, pkg := range pkgStatus[m.Name()].Missing {
			shared.PtermMissing.Printfln("%s %s", m.Icon(), pkg.Name)
			noMissing++
		}

		for _, pkg := range pkgStatus[m.Name()].Removed {
			shared.PtermRemoved.Printfln("%s %s", m.Icon(), pkg.Name)
			noRemoved++
		}
	}

	infoStrings := []string{}
	if noSynced > 0 {
		infoStrings = append(infoStrings, shared.PtermInstalled.Sprintf("%d in sync", noSynced))
	}
	if noUpdated > 0 {
		infoStrings = append(infoStrings, shared.PtermUpdated.Sprintf("%d to update", noUpdated))
	}
	if noMissing > 0 {
		infoStrings = append(infoStrings, shared.PtermMissing.Sprintf("%d to install", noMissing))
	}
	if noRemoved > 0 {
		infoStrings = append(infoStrings, shared.PtermRemoved.Sprintf("%d to remove", noRemoved))
	}

	if len(infoStrings) > 0 {
		fmt.Println("\n" + strings.Join(infoStrings, "   "))
	} else {
		shared.PtermGreen.Printfln("All packages up to date")
	}
	return nil
}
