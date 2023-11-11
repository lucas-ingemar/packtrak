package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
)

type AppFace interface {
	Install(ctx context.Context, apkgs []string, managerName shared.ManagerName, mType manifest.ManifestObjectType, host bool, group string) error
	InstallValidArgsFunc(ctx context.Context, managerName shared.ManagerName, toComplete string, mType manifest.ManifestObjectType) (pkgs []string, err error)
	ListStatus(ctx context.Context, managerNames []shared.ManagerName) (map[shared.ManagerName]shared.DependenciesStatus, map[shared.ManagerName]shared.PackageStatus, error)
	Remove(ctx context.Context, apkgs []string, managerName shared.ManagerName, mType manifest.ManifestObjectType) error
	RemoveValidArgsFunc(ctx context.Context, toComplete string, managerName shared.ManagerName, mType manifest.ManifestObjectType) ([]string, error)
	Sync(ctx context.Context, managerNames []shared.ManagerName) (err error)
	PrintPackageList(depStatus map[shared.ManagerName]shared.DependenciesStatus, pkgStatus map[shared.ManagerName]shared.PackageStatus) error
	//FIXME: Might want to create a typ for manager names
	ListManagers() []shared.ManagerName
	mustDoSudo(ctx context.Context, managers []shared.ManagerName, cmd shared.CommandName) (success bool)
	// GetManifest() manifest.ManifestFace
}

type App struct {
	Managers managers.ManagerFactoryFace
	Manifest manifest.ManifestFace
	State    state.StateFace

	isSudo bool
}

func (a *App) mustDoSudo(ctx context.Context, managerNames []shared.ManagerName, cmd shared.CommandName) (success bool) {
	if a.isSudo {
		return a.isSudo
	}
	managers, err := a.Managers.GetManagers(managerNames)
	if err != nil {
		panic(err)
	}

	pmNames := []string{}
	for _, pm := range managers {
		if lo.Contains(pm.NeedsSudo(), cmd) {
			pmNames = append(pmNames, string(pm.Name()))
		}
	}

	if len(pmNames) == 0 {
		return true
	}

	text := fmt.Sprintf("The following package managers needs sudo privileges to work properly with the '%s' command:\n\n%s\n\nDo you want to grant access? You might need to enter your password", cmd, strings.Join(pmNames, ", "))
	result, _ := pterm.InteractiveContinuePrinter{
		DefaultValueIndex: 0,
		DefaultText:       text,
		TextStyle:         &pterm.ThemeDefault.PrimaryStyle,
		Options:           []string{"y", "n"},
		OptionsStyle:      &pterm.ThemeDefault.SuccessMessageStyle,
		SuffixStyle:       &pterm.ThemeDefault.SecondaryStyle,
		Delimiter:         ": ",
	}.Show()
	if result != "y" {
		return false
	}

	_, err = shared.Command(ctx, "sudo", []string{"echo", ""}, true, os.Stdin)
	if err != nil {
		panic(err)
	}

	a.isSudo = true

	return true
}

func (a *App) ListManagers() []shared.ManagerName {
	return a.Managers.ListManagers()
}

func NewApp(managers managers.ManagerFactoryFace, manifest manifest.ManifestFace, state state.StateFace) *App {
	return &App{
		Managers: managers,
		Manifest: manifest,
		State:    state,
	}
}
