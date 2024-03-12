package flatpak

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
)

type CommandExecutorFace interface {
	ListInstalledPkgs(ctx context.Context, userSpaceInstallation bool) ([]shared.Package, error)
	ListUpdateablePkgs(ctx context.Context, userSpaceInstallation bool) ([]shared.Package, error)
	InstallPkg(ctx context.Context, pkg shared.Package, userSpaceInstallation bool) error
	UpdatePkg(ctx context.Context, pkg shared.Package, userSpaceInstallation bool) error
	RemovePkg(ctx context.Context, pkg shared.Package, userSpaceInstallation bool) error
}

type commandExecutor struct {
}

func (ce commandExecutor) InstallPkg(ctx context.Context, pkg shared.Package, userSpaceInstallation bool) error {
	spaceFlag := "--system"
	if userSpaceInstallation {
		spaceFlag = "--user"
	}

	err := checkNameFormat(pkg.FullName)
	if err != nil {
		return err
	}

	flags := []string{"install", spaceFlag, "--assumeyes"}
	flags = append(flags, strings.Split(pkg.FullName, ":")...)

	_, err = shared.Command(ctx, "flatpak", flags, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ce commandExecutor) UpdatePkg(ctx context.Context, pkg shared.Package, userSpaceInstallation bool) error {
	spaceFlag := "--system"
	if userSpaceInstallation {
		spaceFlag = "--user"
	}

	err := checkNameFormat(pkg.FullName)
	if err != nil {
		return err
	}

	flags := []string{"update", spaceFlag, "--assumeyes", strings.Split(pkg.FullName, ":")[1]}

	_, err = shared.Command(ctx, "flatpak", flags, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ce commandExecutor) RemovePkg(ctx context.Context, pkg shared.Package, userSpaceInstallation bool) error {
	spaceFlag := "--system"
	if userSpaceInstallation {
		spaceFlag = "--user"
	}

	err := checkNameFormat(pkg.FullName)
	if err != nil {
		return err
	}

	flags := []string{"uninstall", spaceFlag, "--assumeyes", strings.Split(pkg.FullName, ":")[1]}

	_, err = shared.Command(ctx, "flatpak", flags, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ce commandExecutor) ListInstalledPkgs(ctx context.Context, userSpaceInstallation bool) (pkgs []shared.Package, err error) {
	spaceFlag := "--system"
	if userSpaceInstallation {
		spaceFlag = "--user"
	}

	stdout, err := shared.Command(ctx, "flatpak", []string{"list", "--columns=origin,application,version", spaceFlag}, false, nil)
	if err != nil {
		return
	}

	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pkg := shared.Package{
			Name:          strings.TrimSpace(fields[1]),
			FullName:      fmt.Sprintf("%s:%s", fields[0], fields[1]),
			Version:       "",
			LatestVersion: "",
			RepoUrl:       "",
		}

		if len(fields) >= 3 {
			pkg.Version = fields[2]
		}
		pkgs = append(pkgs, pkg)
	}

	return
}

func (ce commandExecutor) ListUpdateablePkgs(ctx context.Context, userSpaceInstallation bool) (pkgs []shared.Package, err error) {
	spaceFlag := "--system"
	if userSpaceInstallation {
		spaceFlag = "--user"
	}

	stdout, err := shared.Command(ctx, "flatpak", []string{"remote-ls", "--updates", "--columns=origin,application", spaceFlag}, false, nil)
	if err != nil {
		return
	}

	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pkgs = append(pkgs, shared.Package{
			Name:          "",
			FullName:      fmt.Sprintf("%s:%s", fields[0], fields[1]),
			Version:       "",
			LatestVersion: "",
			RepoUrl:       "",
		})
	}

	return
}
