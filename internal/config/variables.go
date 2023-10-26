package config

import (
	"github.com/adrg/xdg"
	"github.com/spf13/viper"

	"path/filepath"
)

var (
	ConfigDir   string
	DataDir     string
	PackageFile string
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
}

func getViperStringWithDefault(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}
