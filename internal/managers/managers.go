package managers

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/viper"
)

// type ManagerName string

var (
	managersRegistered = []Manager{&Dnf{}, &Go{}}
	PackageManagersOld = []Manager{}
)

type ManagerFactoryFace interface {
	ListManagers() []shared.ManagerName
	GetManager(manager shared.ManagerName) (Manager, error)
	GetManagers(managers []shared.ManagerName) ([]Manager, error)
}

type ManagerFactory struct {
	managers []Manager
}

func (m ManagerFactory) ListManagers() []shared.ManagerName {
	managers := []shared.ManagerName{}
	for _, man := range m.managers {
		managers = append(managers, man.Name())
	}
	return managers
}

func (m ManagerFactory) GetManager(manager shared.ManagerName) (Manager, error) {
	for _, man := range m.managers {
		if man.Name() == manager {
			return man, nil
		}
	}
	return nil, fmt.Errorf("manager '%s' not found", manager)
}

func (m ManagerFactory) GetManagers(managerNames []shared.ManagerName) ([]Manager, error) {
	managers := []Manager{}
	for _, mn := range managerNames {
		manager, err := m.GetManager(mn)
		if err != nil {
			return nil, err
		}
		managers = append(managers, manager)
	}
	return managers, nil
}

type Manager interface {
	Name() shared.ManagerName
	Icon() string
	ShortDesc() string
	LongDesc() string

	NeedsSudo() []shared.CommandName

	InitCheckCmd() error

	GetPackageNames(ctx context.Context, packages []string) []string
	GetDependencyNames(ctx context.Context, deps []string) []string

	InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error)

	AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error)
	AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error)

	ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus shared.DependenciesStatus, err error)
	ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus shared.PackageStatus, err error)

	RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error)
	RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error)

	SyncDependencies(ctx context.Context, depStatus shared.DependenciesStatus) (userWarnings []string, err error)
	SyncPackages(ctx context.Context, packageStatus shared.PackageStatus) (userWarnings []string, err error)
}

func InitManagerConfig() {
	for _, pm := range managersRegistered {
		viper.SetDefault(keyName(pm, "enabled"), true)
	}
}

func InitManagerFactory() (factory ManagerFactory) {
	for _, m := range managersRegistered {
		if viper.GetBool(keyName(m, "enabled")) {
			//FIXME: Here we should also make the init checks
			if err := m.InitCheckCmd(); err != nil {
				shared.PtermWarning.Printfln("Disabling %s manager: %s", m.Name(), err.Error())
				viper.Set(keyName(m, "enabled"), false)
				continue
			}
			factory.managers = append(factory.managers, m)
		}
	}
	return
}

func keyName(m Manager, key string) string {
	return fmt.Sprintf("managers.%s.%s", m.Name(), key)
}
