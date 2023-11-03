package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	depsDevBaseUrl = "https://api.deps.dev/v3alpha/systems"
)

type DepsDevV3alphaWrapper struct {
	PackageKey DepsDevV3alpha   `json:"packageKey"`
	Versions   []DepsDevV3alpha `json:"versions"`
}
type DepsDevV3alpha struct {
	VersionKey DepsDevV3alphaVersionKey `json:"versionKey"`
	IsDefault  bool                     `json:"isDefault"`
}

type DepsDevV3alphaVersionKey struct {
	System  string `json:"system"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func GetDepsDevPackages(manager string, pkg Package) (devPkgs DepsDevV3alphaWrapper, err error) {
	requestURL, err := url.JoinPath(depsDevBaseUrl, manager, "packages", url.QueryEscape(pkg.RepoUrl))
	if err != nil {
		return
	}

	res, err := http.Get(requestURL)
	if err != nil {
		return devPkgs, fmt.Errorf("error making http request: %s", err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return devPkgs, fmt.Errorf("client: could not read response body: %s", err)
	}

	err = json.Unmarshal(resBody, &devPkgs)
	return
}

func GetDepsDevDefaultPackage(manager string, pkg Package) (devPackage DepsDevV3alphaVersionKey, err error) {
	devPkgs, err := GetDepsDevPackages(manager, pkg)
	if err != nil {
		return DepsDevV3alphaVersionKey{}, err
	}
	for _, dpkg := range devPkgs.Versions {
		if dpkg.IsDefault {
			return dpkg.VersionKey, nil
		}
	}
	return DepsDevV3alphaVersionKey{}, errors.New("could not find default version")
}
