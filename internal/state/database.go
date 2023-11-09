package state

import (
	"context"
	"time"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type PackageState struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Manager   string
	Package   string
}

type DependencyState struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	Manager    string
	Dependency string
}

type StateFace interface {
	Begin(ctx context.Context) StateFace
	gorm.TxCommitter
	UpdatePackageState(ctx context.Context, manager string, packages []shared.Package) error
	GetPackageState(ctx context.Context, manager string) (packages []string, err error)
	UpdateDependencyState(ctx context.Context, manager string, deps []shared.Dependency) error
	GetDependencyState(ctx context.Context, manager string) (dependencies []string, err error)
}

type State struct {
	db *gorm.DB
}

func (s State) Begin(ctx context.Context) StateFace {
	tx := s.db.WithContext(ctx).Begin()
	return State{
		db: tx,
	}
}

func (s State) UpdatePackageState(ctx context.Context, manager string, packages []shared.Package) error {
	currentPkgs, err := s.GetPackageState(ctx, manager)
	if err != nil {
		return err
	}

	pkgNames := []string{}
	for _, p := range packages {
		pkgNames = append(pkgNames, p.FullName)
	}

	for _, pkg := range currentPkgs {
		if !lo.Contains(pkgNames, pkg) {
			result := s.db.WithContext(ctx).Where("manager LIKE ? AND package = ?", manager, pkg).Delete(&PackageState{Package: pkg})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, pkg := range pkgNames {
		if !lo.Contains(currentPkgs, pkg) {
			result := s.db.WithContext(ctx).Create(&PackageState{Package: pkg, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

func (s State) GetPackageState(ctx context.Context, manager string) (packages []string, err error) {
	packageStates := []PackageState{}

	result := s.db.WithContext(ctx).Where("manager LIKE ?", manager).Find(&packageStates)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, pkg := range packageStates {
		packages = append(packages, pkg.Package)
	}

	return packages, err
}

func (s State) UpdateDependencyState(ctx context.Context, manager string, deps []shared.Dependency) error {
	currentDeps, err := s.GetDependencyState(ctx, manager)
	if err != nil {
		return err
	}

	depNames := []string{}
	for _, d := range deps {
		depNames = append(depNames, d.FullName)
	}

	for _, dep := range currentDeps {
		if !lo.Contains(depNames, dep) {
			result := s.db.WithContext(ctx).Where("manager LIKE ? AND dependency = ?", manager, dep).Delete(&DependencyState{Dependency: dep})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, dep := range depNames {
		if !lo.Contains(currentDeps, dep) {
			result := s.db.WithContext(ctx).Create(&DependencyState{Dependency: dep, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

func (s State) GetDependencyState(ctx context.Context, manager string) (dependencies []string, err error) {
	depStates := []DependencyState{}

	result := s.db.WithContext(ctx).Where("manager LIKE ?", manager).Find(&depStates)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, dep := range depStates {
		dependencies = append(dependencies, dep.Dependency)
	}

	return dependencies, err
}

func (s State) Commit() error {
	return s.db.Commit().Error
}

func (s State) Rollback() error {
	return s.db.Rollback().Error
}

func NewState(db *gorm.DB) (State, error) {
	err := db.AutoMigrate(&PackageState{}, &DependencyState{})
	return State{db: db}, err
}
