package affected

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type GitChangesProvider struct {
	slnPath string
}

func createGitChangesProvider(slnPath string) GitChangesProvider {
	return GitChangesProvider{slnPath}
}

func (p GitChangesProvider) Get(fromSHA string, toSHA string) ([]string, error) {
	return getChangedFiles(p.slnPath, fromSHA, toSHA)
}

func getChangedFiles(slnPath string, from string, to string) ([]string, error) {
	cmd := exec.Command("git", "diff", from, to, "--name-only")
	cmd.Dir = slnPath

	stdout, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	files := strings.Split(strings.Trim(string(stdout), "\n"), "\n")

	result := []string{}

	gitDir := getGitDir(slnPath)
	slnPrefix, err := filepath.Rel(gitDir, slnPath)
	if err != nil {
		return []string{}, err
	}

	for _, file := range files {
		if strings.HasPrefix(file, slnPrefix) {
			result = append(result, strings.Replace(file, slnPrefix, "", 1))
		}
	}

	return result, nil
}

func getGitDir(slnPath string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = slnPath
	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return strings.Trim(string(stdout), "\n")
}
