package mdnf

import "github.com/lucas-ingemar/mdnf/internal/shared"

func AddPackages(packagesConfig shared.Packages, pkgs []string) (shared.Packages, error) {
	packagesConfig.Global.Packages = append(packagesConfig.Global.Packages, pkgs...)
	return packagesConfig, nil
}
