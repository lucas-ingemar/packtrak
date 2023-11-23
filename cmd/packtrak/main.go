package main

import (
	"os"

	"github.com/lucas-ingemar/packtrak/internal/cmd"
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func main() {
	config.Version = "v0.9.2"
	config.RepoUrl = "https://github.com/lucas-ingemar/packtrak"
	cmd.InitCmd()
	cmd.Execute()
}
