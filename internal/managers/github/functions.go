package github

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
)

func file2Package(filename string) *shared.Package {
	parts := strings.Split(filename, ".")
	if len(parts) != 3 && len(parts) != 4 {
		return nil
	}

	bVersion, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil
	}

	return &shared.Package{
		Name:    fmt.Sprintf("%s/%s", parts[0], parts[1]),
		Version: strings.TrimSpace(string(bVersion)),
	}
}

func package2Filename(pkg shared.Package, fileExtension string) string {
	user, repo, _, _ := url2pkgComponents(pkg.FullName)
	b64Version := base64.StdEncoding.EncodeToString([]byte(pkg.LatestVersion))
	return fmt.Sprintf("%s.%s.%s%s", user, repo, b64Version, fileExtension)
}

func sanitizeGithubUrl(ghUrl string) (sanitizedUrl string, err error) {
	sanitizedUrl = strings.ReplaceAll(ghUrl, "https://", "")
	sanitizedUrl = strings.ReplaceAll(sanitizedUrl, "http://", "")
	urlParts := strings.Split(sanitizedUrl, "/")

	if urlParts[0] != "github.com" {
		return "", errors.New("domain is not github.com")
	}

	if len(urlParts) != 3 {
		return "", errors.New("malformed url")
	}

	subDirFile := strings.Split(urlParts[2], ":")
	if len(subDirFile) != 2 {
		return "", errors.New("no file specified")
	}

	if !strings.Contains(subDirFile[1], "#version#") {
		return "", errors.New("#version# tag not found")
	}

	return
}

func url2pkgComponents(ghUrl string) (user, repo, filePattern string, err error) {
	cmps := strings.Split(ghUrl, ":")
	if len(cmps) != 2 {
		return "", "", "", errors.New("malformed github url")
	}

	urlCmps := strings.Split(cmps[0], "/")
	if len(urlCmps) != 3 {
		return "", "", "", errors.New("malformed github url")
	}
	return urlCmps[1], urlCmps[2], cmps[1], nil
}
