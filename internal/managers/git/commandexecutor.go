package git

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/system"
	"github.com/samber/lo"
)

type CommandExecutorFace interface {
	ListInstalledPkgs(ctx context.Context, folderPath string, includeUnstableReleases bool) ([]shared.Package, error)
	GetRemotePkgMeta(ctx context.Context, pkgUrl string, includeUnstableReleases bool) (shared.Package, error)
	InstallPkg(ctx context.Context, pkg shared.Package, folderPath string) error
	UpdatePkg(ctx context.Context, pkg shared.Package, folderPath string) error
	RemovePkg(ctx context.Context, pkg shared.Package, folderPath string) error
	PkgNameFromUrl(s string) string
	GetBasicPkgInfo(ctx context.Context, pkgNickname string, folderPath string) (shared.Package, error)
}

type commandExecutor struct {
	git system.Git
}

func (c commandExecutor) InstallPkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	repoPath := path.Join(folderPath, strings.ReplaceAll(pkg.Name, "/", "."))
	err := c.git.Clone(ctx, pkg.FullName, repoPath)
	if err != nil {
		return err
	}
	return c.git.Checkout(ctx, repoPath, pkg.LatestVersion)
}

func (c commandExecutor) UpdatePkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	repoPath := path.Join(folderPath, strings.ReplaceAll(pkg.Name, "/", "."))
	err := c.git.Pull(ctx, repoPath)
	if err != nil {
		err = c.RemovePkg(ctx, pkg, folderPath)
		if err != nil {
			return err
		}
		err = c.InstallPkg(ctx, pkg, folderPath)
		if err != nil {
			return err
		}
	}
	return c.git.Checkout(ctx, repoPath, pkg.LatestVersion)
}

func (c commandExecutor) RemovePkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	repoPath := path.Join(folderPath, strings.ReplaceAll(pkg.Name, "/", "."))

	filePath, err := os.Stat(repoPath)
	if err != nil {
		return err
	}

	if !filePath.IsDir() {
		return errors.New("is not a directory")
	}

	return os.RemoveAll(repoPath)
}

func (c commandExecutor) GetRemotePkgMeta(ctx context.Context, pkgUrl string, includeUnstableReleases bool) (pkg shared.Package, err error) {
	pkg.Name = c.PkgNameFromUrl(pkgUrl)
	pkg.RepoUrl = pkgUrl
	pkg.FullName = pkgUrl

	tags, err := c.git.ListRemoteTags(ctx, pkgUrl)
	if err != nil {
		return shared.Package{}, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))
	tags = lo.Filter(tags, func(item string, _ int) bool {
		if item == "latest" {
			return false
		}
		if !includeUnstableReleases && preReleaseTag(item) {
			return false
		}
		return true
	})

	if len(tags) > 0 {
		pkg.LatestVersion = tags[0]
		return
	}

	hash, err := c.git.GetGetRemoteLatestCommitHash(ctx, pkgUrl)
	if err != nil {
		return shared.Package{}, err
	}

	pkg.LatestVersion = hash

	return
}

func (c commandExecutor) ListInstalledPkgs(ctx context.Context, folderPath string, includeUnstableReleases bool) ([]shared.Package, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	pkgs := []shared.Package{}

	for _, e := range files {
		if !e.IsDir() {
			continue
		}
		repoPath := path.Join(folderPath, e.Name())

		remoteUrl, err := c.git.GetRemoteUrl(ctx, repoPath)
		if err != nil {
			return nil, err
		}

		pkg := shared.Package{
			Name:          c.PkgNameFromUrl(remoteUrl),
			FullName:      remoteUrl,
			Version:       "",
			LatestVersion: "",
			RepoUrl:       "",
		}

		tag, err := c.git.GetCurrentTag(ctx, repoPath)
		if err == nil {
			pkg.Version = tag
			pkgs = append(pkgs, pkg)
			continue
		}

		cHash, err := c.git.GetCurrentCommitHash(ctx, repoPath)
		if err != nil {
			return nil, err
		}
		pkg.Version = cHash

		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func (c commandExecutor) GetBasicPkgInfo(ctx context.Context, pkgNickname string, folderPath string) (shared.Package, error) {
	repoPath := path.Join(folderPath, strings.ReplaceAll(pkgNickname, "/", "."))
	remoteUrl, err := c.git.GetRemoteUrl(ctx, repoPath)
	if err != nil {
		return shared.Package{}, err
	}

	pkg := shared.Package{
		Name:          c.PkgNameFromUrl(remoteUrl),
		FullName:      remoteUrl,
		Version:       "",
		LatestVersion: "",
		RepoUrl:       "",
	}
	return pkg, nil
}

func (c commandExecutor) PkgNameFromUrl(s string) string {
	s = strings.TrimSpace(s)
	u, err := url.Parse(s)
	if err != nil {
		return err.Error()
	}
	rString := strings.TrimPrefix(u.Path, "/")
	rString = strings.TrimSuffix(rString, ".git")
	return rString
}

func preReleaseTag(tag string) bool {
	for _, t := range []string{"rc", "alpha", "beta", "pre"} {
		if strings.Contains(tag, t) {
			return true
		}
	}
	return false
}
