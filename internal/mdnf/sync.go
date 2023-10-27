package mdnf

import (
	"context"
	"fmt"

	"github.com/lucas-ingemar/mdnf/internal/dnf"
)

func Sync(ctx context.Context, missingPkgs []string, removedPkgs []string) error {
	if len(missingPkgs) > 0 {
		fmt.Println("")
		err := dnf.Install(ctx, missingPkgs)
		if err != nil {
			return err
		}
	}

	if len(removedPkgs) > 0 {
		fmt.Println("")
		err := dnf.Remove(ctx, removedPkgs)
		if err != nil {
			return err
		}
	}

	return nil
}
