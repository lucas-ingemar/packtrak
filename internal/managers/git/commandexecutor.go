package git

import (
	"context"
	"errors"
	"os"
	"path"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

type CommandExecutorFace interface {
	ListInstalledPkgs(ctx context.Context, folderPath string) ([]shared.Package, error)
	GetRemotePkgMeta(ctx context.Context, pkgUrl string) (shared.Package, error)
	InstallPkg(ctx context.Context, pkg shared.Package, folderPath string) error
	UpdatePkg(ctx context.Context, pkg shared.Package, folderPath string) error
	RemovePkg(ctx context.Context, pkg shared.Package, folderPath string) error
}

type commandExecutor struct {
}

func (c commandExecutor) InstallPkg(ctx context.Context, pkg shared.Package, folderPath string) error {
	_, err := gogit.PlainCloneContext(ctx, path.Join(folderPath, pkg.Name), false, &gogit.CloneOptions{
		URL:               pkg.RepoUrl,
		RecurseSubmodules: gogit.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return err
	}
	return nil
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

func (c commandExecutor) GetRemotePkgMeta(ctx context.Context, pkgUrl string) (pkg shared.Package, err error) {
	pkg.RepoUrl = pkgUrl
	pkg.FullName = pkgUrl
	rem := gogit.NewRemote(memory.NewStorage(), &gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{pkgUrl},
	})

	refs, err := rem.ListContext(ctx, &gogit.ListOptions{
		PeelingOption: gogit.AppendPeeled,
	})

	if err != nil {
		return shared.Package{}, err
	}

	tags := []string{}
	for _, ref := range refs {
		if ref.Name().IsTag() {
			if ref.Name().Short() != "latest" {
				tags = append(tags, strings.ReplaceAll(ref.Name().Short(), "^{}", ""))
			}
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(tags)))
	if len(tags) > 0 {
		pkg.LatestVersion = tags[0]
		return
	}

	r, err := gogit.Clone(memory.NewStorage(), nil, &gogit.CloneOptions{
		URL: pkgUrl,
	})
	if err != nil {
		return shared.Package{}, err
	}

	ref, err := r.Head()
	if err != nil {
		return shared.Package{}, err
	}

	commits, err := r.Log(&gogit.LogOptions{
		From: ref.Hash(),
	})

	if err != nil {
		return shared.Package{}, err
	}

	cm, err := commits.Next()
	if err != nil {
		return shared.Package{}, err
	}

	pkg.LatestVersion = cm.Hash.String()[:7]

	return
}

func (c commandExecutor) ListInstalledPkgs(ctx context.Context, folderPath string) ([]shared.Package, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	pkgs := []shared.Package{}

	for _, e := range files {
		if !e.IsDir() {
			continue
		}
		pkg := shared.Package{
			Name:          e.Name(),
			FullName:      "",
			Version:       "",
			LatestVersion: "",
			RepoUrl:       "",
		}

		r, err := gogit.PlainOpen(path.Join(folderPath, e.Name()))
		if err != nil {
			return nil, err
		}

		remotes, err := r.Remotes()
		if err != nil {
			return nil, err
		}

		if len(remotes) > 0 {
			pkg.FullName = remotes[0].Config().URLs[0]
		}

		tagrefs, err := r.Tags()
		if err != nil {
			return nil, err
		}
		tags := []string{}

		err = tagrefs.ForEach(func(t *plumbing.Reference) error {
			if t.Name().Short() != "latest" {
				tags = append(tags, t.Name().Short())
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		sort.Sort(sort.Reverse(sort.StringSlice(tags)))
		if len(tags) > 0 {
			pkg.Version = tags[0]
			pkgs = append(pkgs, pkg)
			continue
		}

		ref, err := r.Head()
		if err != nil {
			return nil, err
		}

		commits, err := r.Log(&gogit.LogOptions{
			From: ref.Hash(),
		})

		if err != nil {
			return nil, err
		}

		cm, err := commits.Next()
		if err != nil {
			return nil, err
		}

		pkg.Version = cm.Hash.String()[:7]

		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}
