package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func main() {
	sln, err := getSolutionPath(os.Args[1])
	if err != nil {
		panic(err)
	}

	projects, err := parseProjects(sln)
	if err != nil {
		panic(err)
	}

	affected := getAffectedProjects(projects[10:20], projects)

	fmt.Printf("%s", affected)
}

func parseProjects(sln string) ([]Project, error) {
	projectPaths := getProjects(sln)

	cdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(sln)
	if err != nil {
		panic(err)
	}

	result := []Project{}

	for _, projectPath := range projectPaths {
		err = os.Chdir(filepath.Dir(projectPath))
		if err != nil {
			panic(err)
		}

		references := getReferences(sln, projectPath)

		result = append(result, createProject(projectPath, references))

		err = os.Chdir(sln)
		if err != nil {
			panic(err)
		}
	}

	os.Chdir(cdir)

	return result, nil
}

func getAffectedProjects(changed []Project, all []Project) []Project {
	affected := map[string]bool{}

	ring := []Project{}

	for _, changedProject := range changed {
		ring = append(ring, changedProject)
	}

	for len(ring) != 0 {
		currentProject := ring[0]
		ring = ring[1:]

		if affected[currentProject.name] {
			continue
		}

		affected[currentProject.name] = true

		for _, project := range all {
			if currentProject.name == project.name {
				continue
			}

			if contains(project.references, currentProject.name) {
				ring = append(ring, project)
			}
		}
	}

	result := []Project{}

	for k := range affected {
		proj, err := find(all, k)
		if err == nil {
			result = append(result, proj)
		}
	}

	return result
}

func contains(items []string, item string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}

	return false
}

func find(projects []Project, name string) (Project, error) {
	for _, project := range projects {
		if project.name == name {
			return project, nil
		}
	}

	return Project{}, fmt.Errorf("NOT FOUND")
}

func getSolutionPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}

		return filepath.Join(usr.HomeDir, path[2:]), nil
	}

	return path, nil
}

func getProjects(slnPath string) []string {
	cmd := exec.Command("dotnet", "sln", "list")
	cmd.Dir = slnPath

	stdout, err := cmd.Output()
	if err != nil {
		fmt.Printf(err.Error())
	}

	projects := strings.Split(string(stdout), "\n")

	result := []string{}

	for _, project := range projects {
		if strings.Contains(project, ".csproj") {
			result = append(result, project)
		}
	}

	return result
}

func getReferences(slnPath string, projectPath string) []string {
	cmd := exec.Command("dotnet", "list", projectPath, "reference")
	cmd.Dir = slnPath

	stdout, err := cmd.Output()
	if err != nil {
		panic(err.Error())
	}

	references := strings.Split(string(stdout), "\n")

	result := []string{}

	for _, reference := range references {

		if strings.Contains(reference, ".csproj") {
			replaced := strings.ReplaceAll(reference, "\\", "/")
			abs, err := filepath.Abs(replaced)
			if err != nil {
				panic(err)
			}

			if _, err := os.Stat(abs); err != nil {
				continue // no references output from dotnet
			}

			relativeToSln := strings.ReplaceAll(abs, fmt.Sprintf("%s/", slnPath), "")
			result = append(result, relativeToSln)
		}
	}

	return result
}

type Project struct {
	name       string
	references []string
}

func createProject(path string, referencesPaths []string) Project {
	return Project{
		name:       path,
		references: referencesPaths,
	}
}
