package config

import (
	"fmt"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"path/filepath"
)

var (
	ConfigDir   string
	DataDir     string
	PackageFile string
	StateFile   string
)

func init() {
	Refresh()
}

func Refresh() {
	viper.SetEnvPrefix("mdnf")
	viper.AutomaticEnv()
	ConfigDir = getViperStringWithDefault("config_dir", filepath.Join(xdg.ConfigHome, "mdnf"))
	DataDir = getViperStringWithDefault("data_dir", filepath.Join(xdg.DataHome, "mdnf"))
	PackageFile = filepath.Join(ConfigDir, "packages.yml")
	StateFile = filepath.Join(DataDir, "state.yml")

	viper.SetConfigType("yaml")    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(ConfigDir) // path to look for the config file in
	err := viper.ReadInConfig()    // Find and read the config file
	if err != nil {                // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func getViperStringWithDefault(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}
