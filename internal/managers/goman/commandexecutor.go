package goman

import (
	"context"
	"errors"
	"os"
	"path"
	"regexp"

	"github.com/alexellis/go-execute/v2"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

type CommandExecutorFace interface {
	Install(ctx context.Context, pkg shared.Package) error
	Remove(pkg shared.Package) error
	ListInstalled(ctx context.Context) (packages []shared.Package, err error)
	BinPath() (binPath string, err error)
	GetBinaryInfo(ctx context.Context, binaryPath string) (pkg shared.Package, err error)
}

type commandExecutor struct {
}

func (c *commandExecutor) Install(ctx context.Context, pkg shared.Package) error {
	_, err := shared.Command(ctx, "go", []string{"install", pkg.FullName + "@latest"}, false, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *commandExecutor) Remove(pkg shared.Package) error {
	binPath, err := c.BinPath()
	if err != nil {
		return err
	}

	pkgPath := path.Join(binPath, pkg.Name)

	binary, err := os.Stat(pkgPath)
	if err != nil {
		return err
	}

	if binary.IsDir() {
		return errors.New("not a file")
	}

	return os.Remove(pkgPath)
}

func (c *commandExecutor) ListInstalled(ctx context.Context) (packages []shared.Package, err error) {
	binPath, err := c.BinPath()
	if err != nil {
		return nil, err
	}
	binaries, err := os.ReadDir(binPath)
	if err != nil {
		return
	}

	for _, e := range binaries {
		if e.IsDir() {
			continue
		}

		pkg, err := c.GetBinaryInfo(ctx, path.Join(binPath, e.Name()))
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return
}

func (c *commandExecutor) BinPath() (binPath string, err error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", errors.New("GOPATH not found")
	}
	return path.Join(goPath, "bin"), nil
}

func (c *commandExecutor) GetBinaryInfo(ctx context.Context, binaryPath string) (pkg shared.Package, err error) {
	cmd := execute.ExecTask{
		Command:     "go",
		Args:        []string{"version", "-m", binaryPath},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return
	}

	if res.ExitCode != 0 {
		return pkg, errors.New("Non-zero exit code: " + res.Stderr)
	}

	rPath, err := regexp.Compile(`(?m)^\s*path\s*(\S+)$`)
	if err != nil {
		return
	}

	rVersion, err := regexp.Compile(`(?m)^\s*mod\s*(\S+)\s*(\S+)\s*\S+$`)
	if err != nil {
		return
	}

	pathMatches := rPath.FindStringSubmatch(res.Stdout)
	if len(pathMatches) != 2 {
		return pkg, errors.New("could not match path")
	}

	versionMatches := rVersion.FindStringSubmatch(res.Stdout)
	if len(versionMatches) != 3 {
		return pkg, errors.New("could not match version")
	}

	_, name := path.Split(binaryPath)

	return shared.Package{
		Name:     name,
		FullName: pathMatches[1],
		Version:  versionMatches[2],
		RepoUrl:  versionMatches[1],
	}, nil
}
