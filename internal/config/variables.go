package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"path/filepath"
)

var (
	ConfigDir   string
	DataDir     string
	ConfigFile  string
	PackageFile string
	StateFile1  string
	StateFile   string

	ConfigFileExists bool

	DnfEnabled bool
)

const (
	keyDnfEnabled = "dnf.enabled"
)

func init() {
	Refresh()
}

func Refresh() {
	viper.SetEnvPrefix("packtrak")
	viper.AutomaticEnv()

	ConfigDir = filepath.Join(xdg.ConfigHome, "packtrak")
	ConfigFile = filepath.Join(ConfigDir, "config.yaml")
	PackageFile = filepath.Join(ConfigDir, "packages.yaml")

	ConfigFileExists = configFileExists()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(ConfigDir) // path to look for the config file in
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if configFileExists() {
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	DataDir = getViperStringWithDefault("data_dir", filepath.Join(xdg.DataHome, "packtrak"))
	StateFile = filepath.Join(DataDir, "state.db")

	DnfEnabled = getViperBoolWithDefault(keyDnfEnabled, true)
	fmt.Println(DnfEnabled)

	if !configFileExists() {
		err := os.MkdirAll(ConfigDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		err = viper.WriteConfigAs(ConfigFile)
		if err != nil {
			panic(err)
		}
	}
}

func getViperStringWithDefault(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func getViperBoolWithDefault(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func configFileExists() bool {
	_, err := os.Stat(ConfigFile)
	return !os.IsNotExist(err)
}
