package system

import (
	"context"
	"errors"
	"strings"
)

type Git struct {
}

func (g Git) GetRemoteUrl(ctx context.Context, folderPath string) (string, error) {
	u, err := Call().Cmd("git").Args([]string{"remote", "get-url", "origin"}).Cwd(folderPath).Exec(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(u), nil
}

func (g Git) ListTags(ctx context.Context, folderPath string) ([]string, error) {
	res, err := Call().Cmd("git").Args([]string{"tag"}).Cwd(folderPath).Exec(ctx)
	if err != nil {
		return nil, err
	}
	tags := strings.Split(strings.TrimSpace(res), "\n")
	return tags, err
}

func (g Git) GetCurrentTag(ctx context.Context, folderPath string) (string, error) {
	t, err := Call().Cmd("git").Args([]string{"describe", "--tags"}).Cwd(folderPath).Exec(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(t), nil
}

func (g Git) ListCommitHashes(ctx context.Context, folderPath string) ([]string, error) {
	res, err := Call().Cmd("git").Args([]string{"log", "--pretty=oneline", "--abbrev-commit"}).Cwd(folderPath).Exec(ctx)
	if err != nil {
		return nil, err
	}

	var hashes []string
	for _, h := range strings.Split(strings.TrimSpace(res), "\n") {
		hashes = append(hashes, strings.Split(h, " ")[0])
	}
	return hashes, err
}

func (g Git) GetCurrentCommitHash(ctx context.Context, folderPath string) (string, error) {
	hashes, err := g.ListCommitHashes(ctx, folderPath)
	if err != nil {
		return "", err
	}
	if len(hashes) == 0 {
		return "", errors.New("no commits found")
	}
	return hashes[0], nil
}

func (g Git) ListRemoteTags(ctx context.Context, repoUrl string) ([]string, error) {
	res, err := Call().Cmd("git").Args([]string{"ls-remote", "--tags", repoUrl}).Exec(ctx)
	if err != nil {
		return nil, err
	}
	if res == "" {
		return []string{}, nil
	}
	var tags []string
	for _, t := range strings.Split(strings.TrimSpace(res), "\n") {
		t = strings.ReplaceAll(t, "^{}", "")
		t = strings.Split(t, "\t")[1]
		tags = append(tags, strings.TrimPrefix(t, "refs/tags/"))
	}
	return tags, err
}

func (g Git) GetGetRemoteLatestCommitHash(ctx context.Context, repoUrl string) (string, error) {
	res, err := Call().Cmd("git").Args([]string{"ls-remote", repoUrl, "HEAD"}).Exec(ctx)
	if err != nil {
		return "", err
	}
	hashes := strings.Split(strings.TrimSpace(res), "\n")
	if len(hashes) == 0 {
		return "", errors.New("no commits found")
	}
	return hashes[0][:7], nil
}

func (g Git) Clone(ctx context.Context, repoUrl string, folderPath string) error {
	_, err := Call().Cmd("git").Args([]string{"clone", repoUrl, folderPath}).Exec(ctx)
	return err
}

func (g Git) Pull(ctx context.Context, folderPath string) error {
	_, err := Call().Cmd("git").Args([]string{"pull", "origin", "HEAD"}).Cwd(folderPath).Exec(ctx)
	return err
}

func (g Git) Checkout(ctx context.Context, folderPath string, hash string) error {
	_, err := Call().Cmd("git").Args([]string{"checkout", hash}).Cwd(folderPath).Exec(ctx)
	return err
}
