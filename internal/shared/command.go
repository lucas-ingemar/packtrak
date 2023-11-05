package shared

import (
	"context"
	"errors"
	"io"

	"github.com/alexellis/go-execute/v2"
)

func Command(ctx context.Context, command string, args []string, streamStdio bool, stdIn io.Reader) (string, error) {
	cmd := execute.ExecTask{
		Command:     command,
		Args:        args,
		StreamStdio: streamStdio,
		Stdin:       stdIn,
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return "", err
	}

	if res.ExitCode != 0 {
		return "", errors.New("Non-zero exit code: " + res.Stderr)
	}

	return res.Stdout, nil
}
