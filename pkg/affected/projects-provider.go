package affected

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CliProjectsProvider struct {
	slnPath string
}

func createCliProjectsProvider(slnPath string) CliProjectsProvider {
	return CliProjectsProvider{slnPath}
}

func (p CliProjectsProvider) Get() ([]Project, error) {
	return parseProjects(p.slnPath)
}

func parseProjects(sln string) ([]Project, error) {
	projectPaths, err := getProjects(sln)
	if err != nil {
		return []Project{}, err
	}

	cdir, err := os.Getwd()
	if err != nil {
		return []Project{}, err
	}

	err = os.Chdir(sln)
	if err != nil {
		return []Project{}, err
	}

	result := []Project{}

	for _, projectPath := range projectPaths {
		err = os.Chdir(filepath.Dir(projectPath))
		if err != nil {
			return []Project{}, err
		}

		references, err := getReferences(sln, projectPath)
		if err != nil {
			return []Project{}, err
		}

		result = append(result, createProject(projectPath, references))

		err = os.Chdir(sln)
		if err != nil {
			return []Project{}, err
		}
	}

	os.Chdir(cdir)

	return result, nil
}

func getReferences(slnPath string, projectPath string) ([]string, error) {
	cmd := exec.Command("sed", "-n", "s/.*ProjectReference.*Include=\"\\(.*\\)\".*/\\1/p", projectPath)
	cmd.Dir = slnPath

	stdout, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	references := strings.Split(string(stdout), "\n")

	result := []string{}

	for _, reference := range references {

		if strings.Contains(reference, ".csproj") {
			replaced := strings.ReplaceAll(reference, "\\", "/")
			abs, err := filepath.Abs(replaced)
			if err != nil {
				return []string{}, err
			}

			if _, err := os.Stat(abs); err != nil {
				continue // no references
			}

			relativeToSln := strings.ReplaceAll(abs, fmt.Sprintf("%s/", slnPath), "")
			result = append(result, relativeToSln)
		}
	}

	return result, nil
}

func getProjects(slnDirPath string) ([]string, error) {

	slnPath := ""
	filepath.WalkDir(slnDirPath, func(path string, d fs.DirEntry, err error) error {
		extension := filepath.Ext(d.Name())
		if extension == ".sln" {
			slnPath = d.Name()
		}

		return nil
	})

	cmd := exec.Command("sed", "-n", "s/.*Project.*\"\\(.*csproj\\)\".*/\\1/p", slnPath)
	cmd.Dir = slnDirPath
	stdout, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	projects := strings.Split(string(stdout), "\n")

	result := []string{}

	for _, project := range projects {
		if strings.Contains(project, ".csproj") {
			result = append(result, strings.ReplaceAll(project, "\\", "/"))
		}
	}

	return result, nil
}
