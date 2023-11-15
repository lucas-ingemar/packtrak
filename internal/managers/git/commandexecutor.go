package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/alexellis/go-execute/v2"
	gogit "github.com/go-git/go-git/v5"
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
}

type commandExecutor struct {
	git system.Git
}

// FIXME:::::::
// This could be done:
// git clone ...
// git tag -> get the latest good tag
// git checkout *tag*
//
// When updating:
// git describe --tags -> gives the current tag if existing, otherwise error I THINK. YES
// if needs updating:
// git pull origin HEAD
// repeat steps above after clone

func (c commandExecutor) checkoutTag(ctx context.Context, tag string, repoPath string) error {
	cmd := execute.ExecTask{
		Command:     "git",
		Args:        []string{"checkout", "tags/" + tag},
		Cwd:         repoPath,
		Stdin:       nil,
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return errors.New("Non-zero exit code: " + res.Stderr)
	}
	fmt.Println(res.Stdout)
	return nil
}

func (c commandExecutor) InstallPkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	repoPath := path.Join(folderPath, pkg.Name)
	_, err := gogit.PlainCloneContext(ctx, repoPath, false, &gogit.CloneOptions{
		URL:               pkg.RepoUrl,
		RecurseSubmodules: gogit.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return err
	}

	// wt, err := r.Worktree()
	// if err != nil {
	// 	return err
	// }

	// return wt.Checkout(&gogit.CheckoutOptions{
	// 	Hash: plumbing.NewHash(pkg.LatestVersion),
	// })

	return c.checkoutTag(ctx, pkg.LatestVersion, repoPath)
}

func (c commandExecutor) UpdatePkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	r, err := gogit.PlainOpen(path.Join(folderPath, pkg.Name))
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	return w.Pull(&gogit.PullOptions{RemoteName: "origin"})
}

func (c commandExecutor) RemovePkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	pkgPath := path.Join(folderPath, pkg.Name)

	filePath, err := os.Stat(pkgPath)
	if err != nil {
		return err
	}

	if !filePath.IsDir() {
		return errors.New("is not a directory")
	}

	return os.RemoveAll(pkgPath)
}

func (c commandExecutor) GetRemotePkgMeta(ctx context.Context, pkgUrl string, includeUnstableReleases bool) (pkg shared.Package, err error) {
	pkg.Name = pkgNameFromUrl(pkgUrl)
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
			Name:          pkgNameFromUrl(remoteUrl),
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

func preReleaseTag(tag string) bool {
	for _, t := range []string{"rc", "alpha", "beta", "pre"} {
		if strings.Contains(tag, t) {
			return true
		}
	}
	return false
}

func pkgNameFromUrl(s string) string {
	s = strings.TrimSpace(s)
	u, err := url.Parse(s)
	if err != nil {
		return err.Error()
	}
	rString := strings.TrimPrefix(u.Path, "/")
	rString = strings.TrimSuffix(rString, ".git")
	return rString
	// u = strings.TrimSpace(u)
	// u = strings.TrimSuffix(u, ".git")
	// up := strings.Split(u, "/")
	// sort.Sort(sort.Reverse(sort.StringSlice(up)))
	// fmt.Println(up)
	// return fmt.Sprintf("%s/%s", up[1], up[0])
}
