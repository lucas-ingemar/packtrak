package config

import (
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"

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

	CompactPrint   bool
	Groups         []string
	StateRotations int

	AssumeYes *bool
)

const (
	keyCompactPrint   = "compact_print"
	keyGroups         = "groups"
	keyStateRotations = "state_rotations"
	keyVersion        = "_version"
)

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
			log.Fatal().Err(err).Msg("Refresh")
		}
	}

	mustCreateCacheDir()

	DataDir = GetViperStringWithDefault("data_dir", filepath.Join(xdg.DataHome, "packtrak"))
	StateFile = filepath.Join(DataDir, "state.db")

	Groups = getViperStringSliceWithDefault(keyGroups, []string{})
	StateRotations = getViperIntWithDefault(keyStateRotations, 3)

	CompactPrint = getViperBoolWithDefault(keyCompactPrint, false)

	if !configFileExists() {
		err := os.MkdirAll(ConfigDir, os.ModePerm)
		if err != nil {
			log.Fatal().Err(err).Msg("Refresh")
		}
		err = viper.WriteConfigAs(ConfigFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Refresh")
		}
	}

	if semver.Compare(Version, viper.GetString(keyVersion)) > 0 {
		viper.Set(keyVersion, Version)
		err := viper.WriteConfigAs(ConfigFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Refresh")
		}
	}
}

func CheckConfig() {
	if StateRotations > 10 {
		viper.Set(keyStateRotations, 10)
		shared.PtermWarning.Printfln("'%s' (%d) has a limit of 10. Value set to 10.", keyStateRotations, StateRotations)
	}

}

func getViperBoolWithDefault(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func GetViperStringWithDefault(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func getViperIntWithDefault(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

func getViperStringSliceWithDefault(key string, defaultValue []string) []string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringSlice(key)
}

func configFileExists() bool {
	_, err := os.Stat(ConfigFile)
	return !os.IsNotExist(err)
}

func mustCreateCacheDir() {
	err := os.MkdirAll(CacheDir, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("mustCreateCacheDir")
	}
}
