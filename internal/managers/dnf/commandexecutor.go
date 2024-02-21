package dnf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/alexellis/go-execute/v2"
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

type CommandExecutorFace interface {
	InstallPkg(ctx context.Context, pkgs []shared.Package) error
	RemovePkg(ctx context.Context, pkgs []shared.Package) error
	ListInstalledPkgs(ctx context.Context) ([]string, []string, error)
	ListUserInstalledPkgs(ctx context.Context) ([]string, error)
	InstallCm(ctx context.Context, cms string) error
	RemoveCm(ctx context.Context, cm string) error
	ListCm(ctx context.Context) (packages []string, err error)
	InstallCopr(ctx context.Context, copr string) error
	RemoveCopr(ctx context.Context, copr string) error
	ListCoprs(ctx context.Context) ([]string, error)
}

type commandExecutor struct {
	cacheAllInstalled         []string
	cacheAllInstalledVersions []string
	cacheUserInstalled        []string
	cacheCoprs                []string
}

func (d *commandExecutor) yumRepoFolder() string {
	return "/etc/yum.repos.d"
}

func (d *commandExecutor) repoFilePrefix() string {
	return "_packtrak:"
}

func (d *commandExecutor) InstallPkg(ctx context.Context, pkgs []shared.Package) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}

	pkgNames := []string{}
	for _, pkg := range pkgs {
		pkgNames = append(pkgNames, pkg.FullName)
	}

	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "install"}, pkgNames...),
		StreamStdio: true,
		Stdin:       os.Stdin,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return errors.New("Non-zero exit code: " + res.Stderr)
	}

	return nil
}

func (d *commandExecutor) RemovePkg(ctx context.Context, pkgs []shared.Package) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}

	pkgNames := []string{}
	for _, pkg := range pkgs {
		pkgNames = append(pkgNames, pkg.FullName)
	}

	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "remove"}, pkgNames...),
		StreamStdio: true,
		Stdin:       os.Stdin,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return errors.New("Non-zero exit code: " + res.Stderr)
	}

	return nil
}

func (d *commandExecutor) InstallCm(ctx context.Context, cms string) error {
	u, err := url.ParseRequestURI(cms)
	if err != nil {
		return fmt.Errorf("not an url: %s, %s", cms, err)
	}

	repoFileName := path.Join(d.yumRepoFolder(), fmt.Sprintf("%s%s", d.repoFilePrefix(), path.Base(u.Path)))
	cacheRepoFileName := path.Join(config.CacheDir, fmt.Sprintf("%s%s", d.repoFilePrefix(), path.Base(u.Path)))

	res, err := http.Get(cms)
	if err != nil {
		return fmt.Errorf("error making http request: %s", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("client: could not read response body: %s", err)
	}

	err = os.WriteFile(cacheRepoFileName, resBody, 0644)
	if err != nil {
		return fmt.Errorf("yum repo: could not write file %s: %s", repoFileName, err)
	}

	_, err = shared.Command(ctx, "sudo", []string{"chown", "root:root", cacheRepoFileName}, false, nil)
	if err != nil {
		return fmt.Errorf("could not chown %s: %s", cacheRepoFileName, err)
	}

	_, err = shared.Command(ctx, "sudo", []string{"mv", cacheRepoFileName, repoFileName}, false, nil)
	if err != nil {
		return fmt.Errorf("could not move %s: %s", repoFileName, err)
	}

	return nil
}

func (d *commandExecutor) ListCm(ctx context.Context) (packages []string, err error) {
	cms, err := os.ReadDir("/etc/yum.repos.d/")
	if err != nil {
		return []string{}, err
	}

	for _, e := range cms {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), d.repoFilePrefix()) {
			packages = append(packages, strings.ReplaceAll(e.Name(), d.repoFilePrefix(), ""))
		}
	}
	return
}

func (d *commandExecutor) RemoveCm(ctx context.Context, cm string) error {
	u, err := url.ParseRequestURI(cm)
	if err != nil {
		return fmt.Errorf("not an url: %s, %s", cm, err)
	}

	repoFileName := path.Join(d.yumRepoFolder(), fmt.Sprintf("%s%s", d.repoFilePrefix(), path.Base(u.Path)))

	_, err = os.Stat(repoFileName)
	if os.IsNotExist(err) {
		return fmt.Errorf("remove cm: %s, file does not exist", cm)
	}

	_, err = shared.Command(ctx, "sudo", []string{"rm", repoFileName}, false, nil)
	return err
}

func (d *commandExecutor) InstallCopr(ctx context.Context, copr string) error {
	_, err := shared.Command(ctx, "sudo", []string{"dnf", "copr", "enable", copr}, true, os.Stdin)
	return err
}

func (d *commandExecutor) RemoveCopr(ctx context.Context, copr string) error {
	_, err := shared.Command(ctx, "sudo", []string{"dnf", "copr", "remove", copr}, false, nil)
	return err
}

func (d *commandExecutor) ListCoprs(ctx context.Context) ([]string, error) {
	if len(d.cacheCoprs) > 0 {
		return d.cacheCoprs, nil
	}

	ret, err := shared.Command(ctx, "dnf", []string{"copr", "list"}, false, nil)
	if err != nil {
		return nil, err
	}

	lo.ForEach(strings.Split(strings.TrimSpace(ret), "\n"), func(item string, _ int) {
		d.cacheCoprs = append(d.cacheCoprs, strings.TrimSpace(item))
	})

	return d.cacheCoprs, nil
}

func (d *commandExecutor) ListInstalledPkgs(ctx context.Context) ([]string, []string, error) {
	if len(d.cacheAllInstalled) > 0 && len(d.cacheAllInstalledVersions) > 0 {
		return d.cacheAllInstalled, d.cacheAllInstalledVersions, nil
	}

	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"list", "installed"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, nil, err
	}
	if res.ExitCode != 0 {
		return nil, nil, errors.New("Non-zero exit code: " + res.Stderr)
	}

	dnfList := strings.Split(res.Stdout, "\n")
	for _, pkg := range dnfList[1:] {
		d.cacheAllInstalled = append(d.cacheAllInstalled, strings.Split(pkg, ".")[0])
		parts := strings.Fields(pkg)
		if len(parts) > 1 {
			d.cacheAllInstalledVersions = append(d.cacheAllInstalledVersions, parts[1])
		} else {
			d.cacheAllInstalledVersions = append(d.cacheAllInstalledVersions, "")
		}
	}
	return d.cacheAllInstalled, d.cacheAllInstalledVersions, nil
}

func (d *commandExecutor) ListUserInstalledPkgs(ctx context.Context) ([]string, error) {
	if len(d.cacheUserInstalled) > 0 {
		return d.cacheUserInstalled, nil
	}

	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"repoquery", "--userinstalled", "--qf", "%{name} %{version}"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, errors.New("Non-zero exit code: " + res.Stderr)
	}

	dnfList := strings.Split(res.Stdout, "\n")
	for _, pkg := range dnfList {
		d.cacheUserInstalled = append(d.cacheUserInstalled, strings.Split(pkg, " ")[0])
	}

	return d.cacheUserInstalled, nil
}
