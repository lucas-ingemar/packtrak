package app

import (
	"fmt"
	"slices"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
)

func (a *App) PrintPackageList(s status.Status) error {
	noSynced, noUpdated, noMissing, noRemoved := 0, 0, 0, 0

	fmt.Println("\nDependencies:")
	for _, mName := range a.Managers.ListManagers() {
		m, err := a.Managers.GetManager(mName)
		if err != nil {
			return err
		}
		for _, dep := range s.GetDependenciesByStatus(m.Name(), status.StatusSynced) {
			shared.PtermInstalled.Printfln("%s %s", m.Icon(), dep.Name)
			noSynced++
		}

		for _, dep := range s.GetDependenciesByStatus(m.Name(), status.StatusMissing) {
			shared.PtermMissing.Printfln("%s %s", m.Icon(), dep.Name)
			noMissing++
		}

		for _, dep := range s.GetDependenciesByStatus(m.Name(), status.StatusRemoved) {
			shared.PtermRemoved.Printfln("%s %s", m.Icon(), dep.Name)
			noRemoved++
		}
	}

	fmt.Println("\nPackages:")
	if config.CompactPrint {
		ns, nu, nm, nr, err := a.printPackagesStandard(s)
		if err != nil {
			return err
		}
		noSynced += ns
		noUpdated += nu
		noMissing += nm
		noRemoved += nr
	} else {
		ns, nu, nm, nr, err := a.printPackagesEnhanced(s)
		if err != nil {
			return err
		}
		noSynced += ns
		noUpdated += nu
		noMissing += nm
		noRemoved += nr
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

func (a *App) printPackagesEnhanced(s status.Status) (noSynced, noUpdated, noMissing, noRemoved int, err error) {
	syncM, updatedM, missingM, removedM := [][]string{}, [][]string{}, [][]string{}, [][]string{}
	for _, mName := range a.Managers.ListManagers() {
		m, err := a.Managers.GetManager(mName)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusSynced) {
			syncM = append(syncM, []string{shared.PtermInstalled.Sprintf("%s %s", m.Icon(), pkg.Name), shared.PtermGreen.Sprint(pkg.Version)})
			noSynced++
		}

		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusUpdated) {
			updatedM = append(updatedM, []string{shared.PtermUpdated.Sprintf("%s %s", m.Icon(), pkg.Name), shared.PtermYellow.Sprintf("%s -> %s", pkg.Version, pkg.LatestVersion)})
			noUpdated++
		}

		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusMissing) {
			missingM = append(missingM, []string{shared.PtermMissing.Sprintf("%s %s", m.Icon(), pkg.Name), shared.PtermBlue.Sprint(pkg.LatestVersion)})
			noMissing++
		}

		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusRemoved) {
			removedM = append(removedM, []string{shared.PtermRemoved.Sprintf("%s %s", m.Icon(), pkg.Name), shared.PtermRed.Sprint(pkg.Version)})
			noRemoved++
		}
	}

	shared.PtermTablePrinter.WithData(slices.Concat(syncM, updatedM, missingM, removedM)).Render()
	return
}

func (a *App) printPackagesStandard(s status.Status) (noSynced, noUpdated, noMissing, noRemoved int, err error) {
	syncStr, updatedStr, missingStr, removedStr := "", "", "", ""
	for _, mName := range a.Managers.ListManagers() {
		m, err := a.Managers.GetManager(mName)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusSynced) {
			syncStr += shared.PtermInstalled.Sprintfln("%s %s", m.Icon(), pkg.Name)
			noSynced++
		}

		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusUpdated) {
			updatedStr += shared.PtermUpdated.Sprintfln("%s %s %s -> %s", m.Icon(), pkg.Name, pkg.Version, pkg.LatestVersion)
			noUpdated++
		}

		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusMissing) {
			missingStr += shared.PtermMissing.Sprintfln("%s %s", m.Icon(), pkg.Name)
			noMissing++
		}

		for _, pkg := range s.GetPackagesByStatus(m.Name(), status.StatusRemoved) {
			removedStr += shared.PtermRemoved.Sprintfln("%s %s", m.Icon(), pkg.Name)
			noRemoved++
		}
	}

	fmt.Print(syncStr)
	fmt.Print(updatedStr)
	fmt.Print(missingStr)
	fmt.Print(removedStr)
	return
}
