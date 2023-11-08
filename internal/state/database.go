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

// FIXME: I think this can extend gorm.DB. Then we can: tx.UpdatedPackageStatus for a transaction, state.UpdatePackageStatus for hit and leave
type StateFace interface {
	Init() error
	Begin(ctx context.Context) (tx *gorm.DB)
	UpdatePackageState(ctx context.Context, tx *gorm.DB, manager string, packages []shared.Package) error
	GetPackageState(ctx context.Context, tx *gorm.DB, manager string) (packages []string, err error)
	UpdateDependencyState(ctx context.Context, tx *gorm.DB, manager string, deps []shared.Dependency) error
	GetDependencyState(ctx context.Context, tx *gorm.DB, manager string) (dependencies []string, err error)
}

type State struct {
	db *gorm.DB
}

func (s State) Init() error {
	// FIXME: Connect verbose flag, below should be an input to NewState()
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	// 	logger.Config{
	// 		SlowThreshold:             time.Second,   // Slow SQL threshold
	// 		LogLevel:                  logger.Silent, // Log level
	// 		IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
	// 		ParameterizedQueries:      true,          // Don't include params in the SQL log
	// 		Colorful:                  false,         // Disable color
	// 	},
	// )
	// dbInit, err := gorm.Open(sqlite.Open(config.StateFile), &gorm.Config{
	// 	Logger: newLogger,
	// })
	// if err != nil {
	// 	return err
	// }
	err := s.db.AutoMigrate(&PackageState{}, &DependencyState{})
	if err != nil {
		return err
	}

	return nil
}

func (s State) Begin(ctx context.Context) (tx *gorm.DB) {
	return s.db.WithContext(ctx).Begin()
}

// func BeginNoTrans() (tx *gorm.DB) {
// 	return &db
// }

func (s State) UpdatePackageState(ctx context.Context, tx *gorm.DB, manager string, packages []shared.Package) error {
	currentPkgs, err := s.GetPackageState(ctx, tx, manager)
	if err != nil {
		return err
	}

	pkgNames := []string{}
	for _, p := range packages {
		pkgNames = append(pkgNames, p.FullName)
	}

	for _, pkg := range currentPkgs {
		if !lo.Contains(pkgNames, pkg) {
			result := tx.WithContext(ctx).Where("manager LIKE ? AND package = ?", manager, pkg).Delete(&PackageState{Package: pkg})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, pkg := range pkgNames {
		if !lo.Contains(currentPkgs, pkg) {
			result := tx.WithContext(ctx).Create(&PackageState{Package: pkg, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

func (s State) GetPackageState(ctx context.Context, tx *gorm.DB, manager string) (packages []string, err error) {
	packageStates := []PackageState{}

	result := tx.WithContext(ctx).Where("manager LIKE ?", manager).Find(&packageStates)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, pkg := range packageStates {
		packages = append(packages, pkg.Package)
	}

	return packages, err
}

func (s State) UpdateDependencyState(ctx context.Context, tx *gorm.DB, manager string, deps []shared.Dependency) error {
	currentDeps, err := s.GetDependencyState(ctx, tx, manager)
	if err != nil {
		return err
	}

	depNames := []string{}
	for _, d := range deps {
		depNames = append(depNames, d.FullName)
	}

	for _, dep := range currentDeps {
		if !lo.Contains(depNames, dep) {
			result := tx.WithContext(ctx).Where("manager LIKE ? AND dependency = ?", manager, dep).Delete(&DependencyState{Dependency: dep})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, dep := range depNames {
		if !lo.Contains(currentDeps, dep) {
			result := tx.WithContext(ctx).Create(&DependencyState{Dependency: dep, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

func (s State) GetDependencyState(ctx context.Context, tx *gorm.DB, manager string) (dependencies []string, err error) {
	depStates := []DependencyState{}

	result := tx.WithContext(ctx).Where("manager LIKE ?", manager).Find(&depStates)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, dep := range depStates {
		dependencies = append(dependencies, dep.Dependency)
	}

	return dependencies, err
}

func NewState(db *gorm.DB) State {
	return State{
		db: db,
	}
}
