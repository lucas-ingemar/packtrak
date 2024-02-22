package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

type CommandExecutorFace interface {
	ListInstalledPkgs(ctx context.Context, folderPath string) ([]shared.Package, error)
	GetManifestPackages(ctx context.Context, packages []string) ([]shared.Package, error)
	InstallPkg(ctx context.Context, pkg shared.Package, folderPath, binPath string) error
	RemovePkg(ctx context.Context, pkg shared.Package, folderPath, binPath string) error
}

type commandExecutor struct {
	GithubHttpFace
}

func (ce commandExecutor) ListInstalledPkgs(ctx context.Context, folderPath string) (packages []shared.Package, err error) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, nil
	}

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return
	}

	for _, e := range files {
		if e.IsDir() {
			continue
		}
		pkg := file2Package(e.Name())
		if pkg != nil {
			packages = append(packages, *pkg)
		}
	}
	return
}

func (ce commandExecutor) GetManifestPackages(ctx context.Context, packages []string) (pkgObjs []shared.Package, err error) {
	// FIXME: Add Github Access Token ENV
	errs := []error{}
	lo.ForEach(packages, func(pkgFullName string, _ int) {
		pkgFullName, err = sanitizeGithubUrl(pkgFullName)
		if err != nil {
			errs = append(errs, err)
			return
		}
		user, repo, filePattern, err := url2pkgComponents(pkgFullName)
		if err != nil {
			errs = append(errs, err)
			return
		}

		latestVersion, err := ce.GetLatestRelease(ctx, user, repo, filePattern)
		if err != nil {
			errs = append(errs, err)
			return
		}

		pkgObjs = append(pkgObjs, shared.Package{
			Name:          fmt.Sprintf("%s/%s", user, repo),
			FullName:      pkgFullName,
			LatestVersion: latestVersion,
		})
	})
	return pkgObjs, errors.Join(errs...)
}

func (ce commandExecutor) InstallPkg(ctx context.Context, pkg shared.Package, folderPath, binPath string) error {
	newFilename, err := ce.DownloadLatestRelease(ctx, pkg, folderPath)
	if err != nil {
		return err
	}

	err = os.Chmod(newFilename, 0755)
	if err != nil {
		return err
	}

	if binPath != "" {
		_, symlinkName, _, err := url2pkgComponents(pkg.FullName)
		if err != nil {
			return err
		}

		symlinkPath := filepath.Join(binPath, strings.ToLower(symlinkName))
		if _, err := os.Lstat(symlinkPath); err == nil {
			os.Remove(symlinkPath)
		}

		if err = os.Symlink(newFilename, symlinkPath); err != nil {
			return err
		}

	}
	return nil
}

func (ce commandExecutor) RemovePkg(ctx context.Context, pkg shared.Package, folderPath, binPath string) error {
	user, repo, _, err := url2pkgComponents(pkg.FullName)
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(folderPath, fmt.Sprintf("%s.%s*", user, repo)))
	if err != nil {
		return err
	}

	for _, file := range files {
		binary, err := os.Stat(file)
		if err != nil {
			return err
		}

		if binary.IsDir() {
			return errors.New("not a file")
		}

		err = os.Remove(file)
		if err != nil {
			return err
		}
	}

	symlinkPath := filepath.Join(binPath, strings.ToLower(repo))
	if _, err := os.Lstat(symlinkPath); err == nil {
		os.Remove(symlinkPath)
	}

	return nil
}
