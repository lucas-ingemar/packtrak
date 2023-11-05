package cmd

import (
	"fmt"
	"os"

	chigo "github.com/UltiRequiem/chigo/pkg"
	"github.com/common-nighthawk/go-figure"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/spf13/cobra"
)

var PmCmds = map[string]*cobra.Command{}

var rootCmd = &cobra.Command{
	Use:   "packtrak",
	Short: "Managed DNF",
	Long:  chigo.Colorize(figure.NewFigure("packtrak", "speed", true).String()),
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// Do Stuff Here
	// 	fmt.Println("tjof;ljt")
	// },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	packagemanagers.InitPackageManagers()

	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()] = &cobra.Command{
			Use:   pm.Name(),
			Short: fmt.Sprintf("%s en liten beskrivning", pm.Icon()),
			Long:  "En langre beskrivning",
		}
		rootCmd.AddCommand(PmCmds[pm.Name()])
	}

	packagemanagers.MustInitManifest()

	err := state.InitDb()
	if err != nil {
		panic(err)
	}

	// state.Test()

	initInstall()
	initList()
	initRemove()

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
