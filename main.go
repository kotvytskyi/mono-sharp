package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "%s {sln} \n  {sln} string \n\t{sln} solution path\n", os.Args[0])

		flag.PrintDefaults()
	}

	from := flag.String("from", "HEAD", "'from' git commit")
	to := flag.String("to", "HEAD~1", "'to' git commit")

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	sln, err := getSolutionPath(os.Args[1])
	if err != nil {
		panic(err)
	}

	projects, err := parseProjects(sln)
	if err != nil {
		panic(err)
	}

	changedProjects := []Project{}
	files := getChangedFiles(sln, *from, *to)
	for _, file := range files {
		project, err := getFileProject(file, sln, projects)
		if err == nil {
			changedProjects = append(changedProjects, project)
		}
	}

	affected := getAffectedProjects(changedProjects, projects)

	for _, affected := range affected {
		fmt.Println(affected.name)
	}
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

func getChangedFiles(slnPath string, from string, to string) []string {
	cmd := exec.Command("git", "diff", from, to, "--name-only")
	cmd.Dir = slnPath

	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	files := strings.Split(strings.Trim(string(stdout), "\n"), "\n")

	result := []string{}

	gitDir := getGitDir(slnPath)
	slnPrefix, err := filepath.Rel(gitDir, slnPath)
	for _, file := range files {
		if strings.HasPrefix(file, slnPrefix) {
			result = append(result, strings.Replace(file, slnPrefix, "", 1))
		}
	}

	return result
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

// Checks if filepath is relative to any project dir path.
// Returns project or error if not found
func getFileProject(file string, sln string, allProjects []Project) (Project, error) {
	for _, project := range allProjects {
		projectDir := filepath.Join(sln, filepath.Dir(project.name))
		path, err := filepath.Rel(projectDir, filepath.Join(sln, file))
		if err == nil && !strings.Contains(path, "..") {
			return project, nil
		}
	}

	return Project{}, fmt.Errorf("NOT FOUND")
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
