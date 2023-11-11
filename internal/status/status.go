package status

import "github.com/lucas-ingemar/packtrak/internal/shared"

type StatusState string

const (
	StatusSynced  StatusState = "synced"
	StatusUpdated StatusState = "updated"
	StatusMissing StatusState = "missing"
	StatusRemoved StatusState = "removed"
)

type Status struct {
	packages     map[shared.ManagerName]PackageStatus
	dependencies map[shared.ManagerName]DependenciesStatus
}

func (s *Status) AddDependencies(manager shared.ManagerName, status DependenciesStatus) {
	if s.dependencies == nil {
		s.dependencies = map[shared.ManagerName]DependenciesStatus{}
	}
	s.dependencies[manager] = status
}

func (s *Status) AddPackages(manager shared.ManagerName, status PackageStatus) {
	if s.packages == nil {
		s.packages = map[shared.ManagerName]PackageStatus{}
	}
	s.packages[manager] = status
}

func (s Status) CountUpdatedDependencies() (totUpdatedDeps int) {
	for _, dep := range s.dependencies {
		totUpdatedDeps += len(dep.Missing)
		totUpdatedDeps += len(dep.Updated)
		totUpdatedDeps += len(dep.Removed)
	}
	return
}

func (s Status) CountUpdatedPackages() (totUpdatedPkgs int) {
	for _, pkg := range s.packages {
		totUpdatedPkgs += len(pkg.Missing)
		totUpdatedPkgs += len(pkg.Updated)
		totUpdatedPkgs += len(pkg.Removed)
	}
	return
}

func (s Status) GetDependenciesByStatus(manager shared.ManagerName, status StatusState) []shared.Dependency {
	switch status {
	case StatusSynced:
		return s.dependencies[manager].Synced
	case StatusUpdated:
		return s.dependencies[manager].Updated
	case StatusMissing:
		return s.dependencies[manager].Missing
	case StatusRemoved:
		return s.dependencies[manager].Removed
	}
	return []shared.Dependency{}
}

func (s Status) GetPackagesByStatus(manager shared.ManagerName, status StatusState) []shared.Package {
	switch status {
	case StatusSynced:
		return s.packages[manager].Synced
	case StatusUpdated:
		return s.packages[manager].Updated
	case StatusMissing:
		return s.packages[manager].Missing
	case StatusRemoved:
		return s.packages[manager].Removed
	}
	return []shared.Package{}
}

func (s Status) GetDependencies(manager shared.ManagerName) DependenciesStatus {
	return s.dependencies[manager]
}

func (s Status) GetPackages(manager shared.ManagerName) PackageStatus {
	return s.packages[manager]
}

func (s Status) GetUpdatedDependenciesState(managers []shared.ManagerName) map[shared.ManagerName][]shared.Dependency {
	state := map[shared.ManagerName][]shared.Dependency{}
	for _, m := range managers {
		state[m] = []shared.Dependency{}
		state[m] = append(state[m], s.dependencies[m].Synced...)
		state[m] = append(state[m], s.dependencies[m].Updated...)
		state[m] = append(state[m], s.dependencies[m].Missing...)
	}
	return state
}

func (s Status) GetUpdatedPackageState(managers []shared.ManagerName) map[shared.ManagerName][]shared.Package {
	pkgsState := map[shared.ManagerName][]shared.Package{}
	for _, m := range managers {
		pkgsState[m] = []shared.Package{}
		pkgsState[m] = append(pkgsState[m], s.packages[m].Synced...)
		pkgsState[m] = append(pkgsState[m], s.packages[m].Updated...)
		pkgsState[m] = append(pkgsState[m], s.packages[m].Missing...)
	}
	return pkgsState
}

type DependenciesStatus struct {
	Synced  []shared.Dependency
	Updated []shared.Dependency
	Missing []shared.Dependency
	Removed []shared.Dependency
}

type PackageStatus struct {
	Synced  []shared.Package
	Updated []shared.Package
	Missing []shared.Package
	Removed []shared.Package
}
