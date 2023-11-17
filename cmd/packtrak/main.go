package main

import (
	"github.com/lucas-ingemar/packtrak/internal/cmd"
	"github.com/lucas-ingemar/packtrak/internal/config"
)

func main() {
	config.Version = "v1.0.0"
	config.RepoUrl = "https://github.com/lucas-ingemar/packtrak"
	cmd.InitCmd()
	cmd.Execute()
}
