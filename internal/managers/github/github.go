package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/status"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const Name shared.ManagerName = "github"

const (
	packageDirectoryKey = "package_directory"
	binDirectoryKey     = "bin_directory"
	symlinkToBinKey     = "symlink_to_bin"
)

func New() *Github {
	return &Github{}
}

type Github struct {
	pkgDirectory string
	binDirectory string
	symlinkToBin bool
}

func (gh *Github) Name() shared.ManagerName {
	return Name
}

func (gh *Github) Icon() string {
	return ""
}

func (gh *Github) ShortDesc() string {
	return "Manage Github released files"
}

func (gh *Github) LongDesc() string {
	return "Manage Github released files. Download and keep track of artifacts from releases from github"
}

func (gh *Github) NeedsSudo() []shared.CommandName {
	return []shared.CommandName{}
}

func (gh *Github) InitConfig() {
	viper.SetDefault(shared.ConfigKeyName(Name, packageDirectoryKey), "")
	viper.SetDefault(shared.ConfigKeyName(Name, binDirectoryKey), "")
	viper.SetDefault(shared.ConfigKeyName(Name, symlinkToBinKey), false)
}

func (gh *Github) InitCheckCmd() error {
	return nil
}

func (gh *Github) InitCheckConfig() error {
	gh.pkgDirectory = viper.GetString(shared.ConfigKeyName(Name, packageDirectoryKey))
	if gh.pkgDirectory == "" {
		return fmt.Errorf("config '%s' must be set", packageDirectoryKey)
	}

	fInfo, err := os.Stat(gh.pkgDirectory)
	if err != nil {
		return err
	}

	if !fInfo.IsDir() {
		return fmt.Errorf("'%s' is not pointing to a directory", packageDirectoryKey)
	}

	gh.symlinkToBin = viper.GetBool(shared.ConfigKeyName(Name, symlinkToBinKey))
	gh.binDirectory = viper.GetString(shared.ConfigKeyName(Name, binDirectoryKey))
	if gh.symlinkToBin && gh.binDirectory == "" {
		return fmt.Errorf("%s must be set if %s is true", binDirectoryKey, symlinkToBinKey)
	}

	return nil
}

func (gh *Github) GetPackageNames(ctx context.Context, packages []string) []string {
	// FIXME
	return nil
}

func (gh *Github) InstallValidArgs(ctx context.Context, toComplete string, dependencies bool) ([]string, error) {
	// FIXME
	return nil, nil
}

func (gh *Github) AddPackages(ctx context.Context, pkgsToAdd []string) (packagesUpdated []string, userWarnings []string, err error) {
	lo.ForEach(pkgsToAdd, func(pkgName string, _ int) {
		sanitizedPkgName, err := gh.sanitizeGithubUrl(pkgName)
		if err != nil {
			userWarnings = append(userWarnings, err.Error())
			return
		}
		packagesUpdated = append(packagesUpdated, sanitizedPkgName)
	})
	return
}

func (gh *Github) ListPackages(ctx context.Context, packages []string, statePkgs []string) (packageStatus status.PackageStatus, err error) {

	fmt.Println(packages)
	fmt.Println(statePkgs)

	hej, err := gh.url2pkgName("github.com/FreeCAD/FreeCAD:FreeCAD-#version#-Linux-x86_64.AppImage")
	if err != nil {
		return status.PackageStatus{}, err
	}
	return status.PackageStatus{
		Synced: []shared.Package{
			{
				Name:          hej,
				FullName:      "github.com/FreeCAD/FreeCAD:FreeCAD-#version#-Linux-x86_64.AppImage",
				Version:       "",
				LatestVersion: "",
				RepoUrl:       "",
			},
		},
		Updated: []shared.Package{},
		Missing: []shared.Package{},
		Removed: []shared.Package{},
	}, err
	// FIXME
	return
}

func (gh *Github) RemovePackages(ctx context.Context, allPkgs []string, pkgsToRemove []string) (packagesToRemove []string, userWarnings []string, err error) {
	// FIXME
	return
}

func (gh *Github) SyncPackages(ctx context.Context, packageStatus status.PackageStatus) (userWarnings []string, err error) {
	// FIXME
	return
}

func (gh *Github) sanitizeGithubUrl(ghUrl string) (sanitizedUrl string, err error) {
	sanitizedUrl = strings.ReplaceAll(ghUrl, "https://", "")
	sanitizedUrl = strings.ReplaceAll(sanitizedUrl, "http://", "")
	urlParts := strings.Split(sanitizedUrl, "/")

	if urlParts[0] != "github.com" {
		return "", errors.New("domain is not github.com")
	}

	if len(urlParts) != 3 {
		return "", errors.New("malformed url")
	}

	subDirFile := strings.Split(urlParts[2], ":")
	if len(subDirFile) != 2 {
		return "", errors.New("no file specified")
	}

	if !strings.Contains(subDirFile[1], "#version#") {
		return "", errors.New("#version# tag not found")
	}

	return
}

func (gh *Github) url2pkgName(ghUrl string) (string, error) {
	cmps := strings.Split(ghUrl, ":")
	if len(cmps) != 2 {
		return "", errors.New("malformed github url")
	}

	urlCmps := strings.Split(cmps[0], "/")
	if len(urlCmps) != 3 {
		return "", errors.New("malformed github url")
	}
	return fmt.Sprintf("%s/%s", urlCmps[1], urlCmps[2]), nil
}

func (gh *Github) GetDependencyNames(ctx context.Context, deps []string) []string {
	return nil
}

func (gh *Github) AddDependencies(ctx context.Context, depsToAdd []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (gh *Github) ListDependencies(ctx context.Context, deps []string, stateDeps []string) (depStatus status.DependenciesStatus, err error) {
	return
}

func (gh *Github) RemoveDependencies(ctx context.Context, allDeps []string, depsToRemove []string) (depsUpdated []string, userWarnings []string, err error) {
	return
}

func (gh *Github) SyncDependencies(ctx context.Context, depStatus status.DependenciesStatus) (userWarnings []string, err error) {
	return
}
