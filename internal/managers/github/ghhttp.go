package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/tidwall/gjson"
)

type GithubHttpFace interface {
	GetLatestRelease(ctx context.Context, user, repo, filePattern string) (version string, err error)
	DownloadLatestRelease(ctx context.Context, pkg shared.Package, targetFolder string) (newFilename string, err error)
}

type GithubHttp struct {
}

func (g GithubHttp) GetLatestRelease(ctx context.Context, user, repo, filePattern string) (version string, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	version = gjson.GetBytes(body, "tag_name").Str
	if version == "" {
		return "", fmt.Errorf("could not find version for %s/%s", user, repo)
	}
	return
}

func (g GithubHttp) DownloadLatestRelease(ctx context.Context, pkg shared.Package, targetFolder string) (newFilename string, err error) {
	user, repo, filePattern, err := url2pkgComponents(pkg.FullName)
	if err != nil {
		return
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	filename := strings.ReplaceAll(filePattern, "#version#", pkg.LatestVersion)

	binaryUrl := gjson.GetBytes(body, fmt.Sprintf(`assets.#(name="%s").browser_download_url`, filename)).Str
	if binaryUrl == "" {
		return "", errors.New("could not find latest release url")
	}

	newFilename = filepath.Join(targetFolder, package2Filename(pkg, filepath.Ext(binaryUrl)))
	out, err := os.Create(newFilename)
	if err != nil {
		return
	}
	defer out.Close()
	BinaryResp, err := http.Get(binaryUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, BinaryResp.Body)

	return newFilename, err
}
