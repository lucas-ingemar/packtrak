package state

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lucas-ingemar/packtrak/internal/config"
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

	db.AutoMigrate(&PackageState{})

	return nil

	// db.Create(&PackageState{
	// 	Manager: "dnf",
	// 	Package: "pack1",
	// })

	// db.Create(&PackageState{
	// 	Manager: "git",
	// 	Package: "pack2",
	// })
	// p := []PackageState{}
	// db.Find(&p)
	// b, _ := json.Marshal(p)
	// fmt.Println(string(b))
}

func Begin() (tx *gorm.DB) {
	return db.Begin()
}

func BeginNoTrans() (tx *gorm.DB) {
	return &db
}

func UpdatePackageState(tx *gorm.DB, manager string, packages []string) error {
	currentPkgs, err := GetPackageState(tx, manager)
	if err != nil {
		return err
	}

	for _, pkg := range currentPkgs {
		if !lo.Contains(packages, pkg) {
			result := tx.Where("manager LIKE ? AND package = ?", manager, pkg).Delete(&PackageState{Package: pkg})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, pkg := range packages {
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

func Test() {
	tx := Begin()
	defer tx.Rollback()

	// p := []PackageState{}
	// db.Find(&p)
	// b, _ := json.Marshal(p)
	// fmt.Println(string(b))

	err := UpdatePackageState(tx, "go", []string{})
	if err != nil {
		panic(err)
	}
	pkgs, err := GetPackageState(tx, "go")
	if err != nil {
		panic(err)
	}
	fmt.Println(pkgs)

	// tx.Delete(&[]PackageState{{Manager: "dnf"}})
	// tx.Create(&PackageState{
	// 	Manager: "go",
	// 	Package: "pack1",
	// })

	res := tx.Commit()
	fmt.Println(res.Error)
}
