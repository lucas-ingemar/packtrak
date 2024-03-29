package cmd

import (
	"fmt"
	"os"

	chigo "github.com/UltiRequiem/chigo/pkg"
	"github.com/common-nighthawk/go-figure"
	"github.com/glebarez/sqlite"
	"github.com/lucas-ingemar/packtrak/internal/app"
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var PmCmds = map[shared.ManagerName]*cobra.Command{}

var rootCmd = &cobra.Command{
	Use:   "packtrak",
	Short: "Track your package managers",
	Long:  chigo.Colorize(figure.NewFigure("packtrak", "speed", true).String()),
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func InitCmd() {
	if shared.IsSudo() {
		shared.PtermWarning.Println("This command can't be run under sudo. You will be prompted later if sudo is needed.")
		os.Exit(1)
	}

	managers.InitManagerConfig()
	config.Refresh()
	mf := managers.InitManagerFactory(managers.ManagersRegistered, true)

	for _, mName := range mf.ListManagers() {
		manager, err := mf.GetManager(mName)
		if err != nil {
			log.Fatal().Err(err).Msg("InitCmd")
		}
		PmCmds[manager.Name()] = &cobra.Command{
			Use:   string(manager.Name()),
			Short: manager.ShortDesc(),
			Long:  manager.LongDesc(),
		}
		rootCmd.AddCommand(PmCmds[manager.Name()])
	}

	// FIXME: Connect verbose flag, below should be an input to NewState()
	// FIXME: Connect zerolog
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

	m, err := manifest.InitManifest()
	if err != nil {
		log.Fatal().Err(err).Msg("InitCmd")
	}

	db, err := gorm.Open(sqlite.Open(config.StateFile), &gorm.Config{
		// FIXME
		// Logger: newLogger,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("InitCmd")
	}

	s, err := state.NewState(db)
	if err != nil {
		log.Fatal().Err(err).Msg("InitCmd")
	}

	a := app.NewApp(mf, &m, s)
	initInstall(a)
	initList(a)
	initRemove(a)
	initSync(a)

	config.CheckConfig()

	config.AssumeYes = rootCmd.PersistentFlags().BoolP("assumeyes", "y", false, "Automatically answer yes for all questions")

	// cobra.OnInitialize(initConfig)
	// rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "Author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	// rootCmd.PersistentFlags().StringVarP(&projectBase, "projectbase", "b", "", "base project directory eg. github.com/spf13/")
	// rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "Name of license for the project (can provide `licensetext` in config)")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("projectbase", rootCmd.PersistentFlags().Lookup("projectbase"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")
}
