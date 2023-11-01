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
	UpdatedAt time.Time
	Manager   string
	Package   string
}

func InitDb() error {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
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

func UpdatePackageState(tx *gorm.DB, manager string, pkgsAdd, pkgsRemove []string) error {
	currentPkgs, err := GetPackageState(tx, manager)
	if err != nil {
		return err
	}

	for _, pkg := range pkgsRemove {
		if lo.Contains(currentPkgs, pkg) {
			fmt.Println("remove: ", pkg)
			result := tx.Where("manager LIKE ? AND package = ?", manager, pkg).Delete(&PackageState{Package: pkg})
			if result.Error != nil {
				return result.Error
			}
		}
	}

	for _, pkg := range pkgsAdd {
		if !lo.Contains(currentPkgs, pkg) {
			fmt.Println("add: ", pkg)
			result := tx.Create(&PackageState{Package: pkg, Manager: manager})
			if result.Error != nil {
				return result.Error
			}
		}
	}
	return nil
	// pkgsR := []PackageState{}
	// for _, pkg := range pkgsRemove {
	// 	pkgsR = append(pkgsR, PackageState{
	// 		Manager: manager,
	// 		Package: pkg,
	// 	})
	// }

	// pkgsA := []PackageState{}
	// for _, pkg := range pkgsAdd {
	// 	pkgsA = append(pkgsA, PackageState{
	// 		Manager: manager,
	// 		Package: pkg,
	// 	})
	// }

	// result := tx.Create(&pkgsA)
	// return result.Error
	// if result.Error != nil {
	// 	return result.Error
	// }

	// return tx.Commit().Error
}

func GetPackageState(tx *gorm.DB, manager string) (packages []string, err error) {
	packageStates := []PackageState{}

	result := tx.Where("manager LIKE ?", manager).Find(&packageStates)
	fmt.Println(result.RowsAffected)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, pkg := range packageStates {
		fmt.Println(pkg.Package)
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

	err := UpdatePackageState(tx, "go", []string{"p3", "p2"}, []string{"p1"})
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
