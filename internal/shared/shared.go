package shared

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/samber/lo"
)

var (
	isSudo bool
)

func IsSudo() bool {
	if os.Getenv("SUDO_UID") != "" && os.Getenv("SUDO_GID") != "" && os.Getenv("SUDO_USER") != "" {
		return true
	}
	return false
}

func GetPackage(name string, packages []Package) (Package, error) {
	for _, pkg := range packages {
		if pkg.Name == name {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("package %s not found", name)
}

func MustDoSudo(ctx context.Context, packageManagers []PackageManager, cmd CommandName) (success bool) {
	if isSudo {
		return isSudo
	}
	pmNames := []string{}
	for _, pm := range packageManagers {
		if lo.Contains(pm.NeedsSudo(), cmd) {
			pmNames = append(pmNames, pm.Name())
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

	_, err := Command(ctx, "sudo", []string{"echo", ""}, true, os.Stdin)
	if err != nil {
		panic(err)
	}

	isSudo = true

	return true
}
