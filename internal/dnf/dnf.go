package dnf

import (
	"context"
	"errors"
	"os"
	"strings"

	execute "github.com/alexellis/go-execute/v2"
)

func ListInstalled(ctx context.Context) ([]string, error) {
	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"list", "installed"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, errors.New("Non-zero exit code: " + res.Stderr)
	}
	// fmt.Printf("stdout: %s, stderr: %s, exit-code: %d\n", res.Stdout, res.Stderr, res.ExitCode)

	dnfList := strings.Split(res.Stdout, "\n")

	return dnfList[1:], nil
}

func ListAvailable(ctx context.Context, filter string) ([]string, error) {
	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"list", "--available", filter + "*"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if res.ExitCode != 0 {
		return nil, errors.New("Non-zero exit code: " + res.Stderr)
	}

	dnfList := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	for idx, line := range dnfList {
		if strings.Contains(line, "Available Packages") {
			dnfList = dnfList[idx+1:]
			break
		}
	}

	pkgs := []string{}

	for _, pkg := range dnfList {
		pkgs = append(pkgs, strings.Split(pkg, ".")[0])
	}

	return pkgs, nil
}

func Install(ctx context.Context, pkgs []string) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}
	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "install"}, pkgs...),
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

func Remove(ctx context.Context, pkgs []string) error {
	if len(pkgs) == 0 {
		return errors.New("no packages provided")
	}
	cmd := execute.ExecTask{
		Command:     "sudo",
		Args:        append([]string{"dnf", "--color=always", "remove"}, pkgs...),
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
