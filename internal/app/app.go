package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/managers"
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type AppFace interface {
	Install(ctx context.Context, apkgs []string, managerName shared.ManagerName, mType manifest.ManifestObjectType, host bool, group string) error
	InstallValidArgsFunc(ctx context.Context, managerName shared.ManagerName, toComplete string, mType manifest.ManifestObjectType) (pkgs []string, err error)
	ListStatus(ctx context.Context, managerNames []shared.ManagerName) (status.Status, error)
	Remove(ctx context.Context, apkgs []string, managerName shared.ManagerName, mType manifest.ManifestObjectType) error
	RemoveValidArgsFunc(ctx context.Context, toComplete string, managerName shared.ManagerName, mType manifest.ManifestObjectType) ([]string, error)
	Sync(ctx context.Context, managerNames []shared.ManagerName) (err error)
	PrintPackageList(s status.Status) error
	ListManagers() []shared.ManagerName
	mustDoSudo(ctx context.Context, managers []shared.ManagerName, cmd shared.CommandName) (success bool)
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
		log.Fatal().Err(err).Msg("mustDoSudo")
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

	if !*config.AssumeYes {
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
	}

	_, err = shared.Command(ctx, "sudo", []string{"echo", ""}, true, os.Stdin)
	if err != nil {
		log.Fatal().Err(err).Msg("mustDoSudo")
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
