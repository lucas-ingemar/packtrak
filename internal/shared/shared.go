package shared

import (
	"os"
)

func IsSudo() bool {
	if os.Getenv("SUDO_UID") != "" && os.Getenv("SUDO_GID") != "" && os.Getenv("SUDO_USER") != "" {
		return true
	}
	return false
}
