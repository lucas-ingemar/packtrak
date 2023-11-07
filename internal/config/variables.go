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
	ConfigDir    string
	DataDir      string
	CacheDir     string
	ConfigFile   string
	ManifestFile string
	StateFile    string

	ConfigFileExists bool

	Version string
	RepoUrl string

	DnfEnabled bool
	Groups     []string
)

const (
	// keyDnfEnabled = "dnf.enabled"
	keyGroups = "groups"
)

// func init() {
// 	Refresh()
// }

// type Hej struct {
// 	// Type  types.String
// 	Name  string
// 	Value interface{}
// }

func Refresh() {
	viper.SetEnvPrefix("packtrak")
	viper.AutomaticEnv()

	ConfigDir = filepath.Join(xdg.ConfigHome, "packtrak")
	ConfigFile = filepath.Join(ConfigDir, "config.yaml")
	ManifestFile = filepath.Join(ConfigDir, "manifest.yaml")

	CacheDir = filepath.Join(xdg.CacheHome, "packtrak")

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
	// hh := Hej{
	// 	Name:  "test",
	// 	Value: []string{"hej"},
	// }

	// // rs := reflect.SliceOf(reflect.String)
	// // bb := hh.Value.(hh.Type.Underlying)
	// fmt.Println(reflect.ValueOf(hh.Value).Kind() == reflect.SliceOf(reflect.TypeOf("")).Kind())
	// panic("hej")

	mustCreateCacheDir()

	DataDir = getViperStringWithDefault("data_dir", filepath.Join(xdg.DataHome, "packtrak"))
	StateFile = filepath.Join(DataDir, "state.db")

	Groups = getViperStringSliceWithDefault(keyGroups, []string{})

	// DnfEnabled = getViperBoolWithDefault(keyDnfEnabled, true)
	// fmt.Println(DnfEnabled)

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

func getViperStringSliceWithDefault(key string, defaultValue []string) []string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringSlice(key)
}

func getViperBoolWithDefault(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func configFileExists() bool {
	_, err := os.Stat(ConfigFile)
	return !os.IsNotExist(err)
}

func mustCreateCacheDir() {
	err := os.MkdirAll(CacheDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
