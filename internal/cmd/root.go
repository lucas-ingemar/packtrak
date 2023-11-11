package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	chigo "github.com/UltiRequiem/chigo/pkg"
	"github.com/common-nighthawk/go-figure"
	"github.com/glebarez/sqlite"
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var PmCmds = map[managers.ManagerName]*cobra.Command{}

var rootCmd = &cobra.Command{
	Use:   "packtrak",
	Short: "Managed DNF",
	Long:  chigo.Colorize(figure.NewFigure("packtrak", "speed", true).String()),
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// Do Stuff Here
	// 	fmt.Println("tjof;ljt")
	// },
}

func Hej() *cobra.Command {
	return rootCmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if shared.IsSudo() {
		shared.PtermWarning.Println("This command can't be run under sudo. You will be prompted later if sudo is needed.")
		os.Exit(1)
	}

	managers.InitManagerConfig()
	config.Refresh()
	mf := managers.InitManagerFactory()

	for _, mName := range mf.ListManagers() {
		manager, err := mf.GetManager(mName)
		if err != nil {
			panic(err)
		}
		PmCmds[manager.Name()] = &cobra.Command{
			Use:   string(manager.Name()),
			Short: manager.ShortDesc(),
			Long:  manager.LongDesc(),
		}
		rootCmd.AddCommand(PmCmds[manager.Name()])
	}

	// err := state.InitDb()
	// if err != nil {
	// 	panic(err)
	// }

	// state.Test()

	// FIXME: Connect verbose flag, below should be an input to NewState()
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

	m, err := manifest.InitManifest()
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(sqlite.Open(config.StateFile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	s, err := state.NewState(db)
	if err != nil {
		panic(err)
	}

	a := app.NewApp(mf, &m, s)
	initInstall(a)
	initList(a)
	initRemove(a)
	initSync(a)

	config.CheckConfig()

	cobra.OnInitialize(initConfig)
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	// rootCmd.PersistentFlags().StringVarP(&projectBase, "projectbase", "b", "", "base project directory eg. github.com/spf13/")
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "Author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "Name of license for the project (can provide `licensetext` in config)")
	// rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("projectbase", rootCmd.PersistentFlags().Lookup("projectbase"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")
}

func initConfig() {
	// FIXME: Add cfg file with secret

	// // Don't forget to read config either from cfgFile or from home directory!
	// if cfgFile != "" {
	// 	// Use config file from the flag.
	// 	viper.SetConfigFile(cfgFile)
	// } else {
	// 	// Find home directory.
	// 	home, err := homedir.Dir()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}

	// 	// Search config in home directory with name ".cobra" (without extension).
	// 	viper.AddConfigPath(home)
	// 	viper.SetConfigName(".cobra")
	// }

	// if err := viper.ReadInConfig(); err != nil {
	// 	fmt.Println("Can't read config:", err)
	// 	os.Exit(1)
	// }
}
