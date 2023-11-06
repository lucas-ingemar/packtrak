package state

import (
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db gorm.DB
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

func InitDb() error {
	// FIXME: Connect verbose flag
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)
	dbInit, err := gorm.Open(sqlite.Open(config.StateFile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}
	db = *dbInit

	err = db.AutoMigrate(&PackageState{}, &DependencyState{})
	if err != nil {
		return err
	}

	return nil
}

func Begin() (tx *gorm.DB) {
	return db.Begin()
}

func BeginNoTrans() (tx *gorm.DB) {
	return &db
}

func UpdatePackageState(tx *gorm.DB, manager string, packages []shared.Package) error {
	currentPkgs, err := GetPackageState(tx, manager)
	if err != nil {
		return err
	}

	pkgNames := []string{}
	for _, p := range packages {
		pkgNames = append(pkgNames, p.FullName)
	}

	for _, pkg := range currentPkgs {
		if !lo.Contains(pkgNames, pkg) {
			result := tx.Where("manager LIKE ? AND package = ?", manager, pkg).Delete(&PackageState{Package: pkg})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, pkg := range pkgNames {
		if !lo.Contains(currentPkgs, pkg) {
			result := tx.Create(&PackageState{Package: pkg, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

func GetPackageState(tx *gorm.DB, manager string) (packages []string, err error) {
	packageStates := []PackageState{}

	result := tx.Where("manager LIKE ?", manager).Find(&packageStates)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, pkg := range packageStates {
		packages = append(packages, pkg.Package)
	}

	return packages, err
}

func UpdateDependencyState(tx *gorm.DB, manager string, deps []shared.Dependency) error {
	currentDeps, err := GetDependencyState(tx, manager)
	if err != nil {
		return err
	}

	depNames := []string{}
	for _, d := range deps {
		depNames = append(depNames, d.FullName)
	}

	for _, dep := range currentDeps {
		if !lo.Contains(depNames, dep) {
			result := tx.Where("manager LIKE ? AND dependency = ?", manager, dep).Delete(&DependencyState{Dependency: dep})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, dep := range depNames {
		if !lo.Contains(currentDeps, dep) {
			result := tx.Create(&DependencyState{Dependency: dep, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
}

func GetDependencyState(tx *gorm.DB, manager string) (dependencies []string, err error) {
	depStates := []DependencyState{}

	result := tx.Where("manager LIKE ?", manager).Find(&depStates)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, dep := range depStates {
		dependencies = append(dependencies, dep.Dependency)
	}

	return dependencies, err
}

// func Test() {
// 	tx := Begin()
// 	defer tx.Rollback()

// 	// p := []PackageState{}
// 	// db.Find(&p)
// 	// b, _ := json.Marshal(p)
// 	// fmt.Println(string(b))

// 	err := UpdatePackageState(tx, "go", []string{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	pkgs, err := GetPackageState(tx, "go")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(pkgs)

// 	// tx.Delete(&[]PackageState{{Manager: "dnf"}})
// 	// tx.Create(&PackageState{
// 	// 	Manager: "go",
// 	// 	Package: "pack1",
// 	// })

// 	res := tx.Commit()
// 	fmt.Println(res.Error)
// }