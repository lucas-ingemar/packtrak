package core

// import (
// 	"fmt"
// 	"strings"

// 	"github.com/lucas-ingemar/packtrak/internal/managers"
// 	"github.com/lucas-ingemar/packtrak/internal/shared"
// )

// func PrintPackageList(depStatus map[managers.ManagerName]shared.DependenciesStatus, pkgStatus map[managers.ManagerName]shared.PackageStatus) {
// 	noSynced, noUpdated, noMissing, noRemoved := 0, 0, 0, 0

// 	fmt.Println("\nDependencies:")
// 	for _, pm := range managers.PackageManagers {
// 		for _, dep := range depStatus[pm.Name()].Synced {
// 			shared.PtermInstalled.Printfln("%s %s", pm.Icon(), dep.Name)
// 			noSynced++
// 		}

// 		for _, dep := range depStatus[pm.Name()].Missing {
// 			shared.PtermMissing.Printfln("%s %s", pm.Icon(), dep.Name)
// 			noMissing++
// 		}

// 		for _, dep := range depStatus[pm.Name()].Removed {
// 			shared.PtermRemoved.Printfln("%s %s", pm.Icon(), dep.Name)
// 			noRemoved++
// 		}
// 	}

// 	fmt.Println("\nPackages:")
// 	for _, pm := range managers.PackageManagers {
// 		for _, pkg := range pkgStatus[pm.Name()].Synced {
// 			shared.PtermInstalled.Printfln("%s %s", pm.Icon(), pkg.Name)
// 			noSynced++
// 		}

// 		for _, pkg := range pkgStatus[pm.Name()].Updated {
// 			shared.PtermUpdated.Printfln("%s %s %s -> %s", pm.Icon(), pkg.Name, pkg.Version, pkg.LatestVersion)
// 			noUpdated++
// 		}

// 		for _, pkg := range pkgStatus[pm.Name()].Missing {
// 			shared.PtermMissing.Printfln("%s %s", pm.Icon(), pkg.Name)
// 			noMissing++
// 		}

// 		for _, pkg := range pkgStatus[pm.Name()].Removed {
// 			shared.PtermRemoved.Printfln("%s %s", pm.Icon(), pkg.Name)
// 			noRemoved++
// 		}
// 	}

// 	infoStrings := []string{}
// 	if noSynced > 0 {
// 		infoStrings = append(infoStrings, shared.PtermInstalled.Sprintf("%d in sync", noSynced))
// 	}
// 	if noUpdated > 0 {
// 		infoStrings = append(infoStrings, shared.PtermUpdated.Sprintf("%d to update", noUpdated))
// 	}
// 	if noMissing > 0 {
// 		infoStrings = append(infoStrings, shared.PtermMissing.Sprintf("%d to install", noMissing))
// 	}
// 	if noRemoved > 0 {
// 		infoStrings = append(infoStrings, shared.PtermRemoved.Sprintf("%d to remove", noRemoved))
// 	}

// 	if len(infoStrings) > 0 {
// 		fmt.Println("\n" + strings.Join(infoStrings, "   "))
// 	} else {
// 		shared.PtermGreen.Printfln("All packages up to date")
// 	}
// }
