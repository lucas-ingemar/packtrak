package dnf

import (
	"context"
	"errors"
	"strings"

	execute "github.com/alexellis/go-execute/v2"
)

func ListInstalled() ([]string, error) {
	cmd := execute.ExecTask{
		Command:     "dnf",
		Args:        []string{"list", "installed"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(context.Background())
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
