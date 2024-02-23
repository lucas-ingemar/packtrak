package flatpak

import (
	"fmt"
	"strings"
)

func checkNameFormat(fullName string) error {
	genericError := fmt.Errorf("wrong format: '%s'. Should be 'remote:application_id', e.g 'flathub:com.slack.Slack'", fullName)
	cmps := strings.Split(fullName, ":")
	if len(cmps) != 2 {
		return genericError
	}

	if len(strings.Split(cmps[1], ".")) == 0 {
		return genericError
	}

	return nil
}
