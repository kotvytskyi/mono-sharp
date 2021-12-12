package affected

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Project struct {
	Path            string
	ReferencesPaths []string
}

func createProject(path string, referencesPaths []string) Project {
	return Project{path, referencesPaths}
}

type ProjectsProvider interface {
	Get() ([]Project, error)
}

type ChangesProvider interface {
	Get(fromSHA string, toSHA string) ([]string, error)
}

type Affected struct {
	slnPath          string
	changesProvider  ChangesProvider
	projectsProvider ProjectsProvider
}

func CreateAffected(slnDirPath string) (Affected, error) {
	if slnDirPath == "" {
		return Affected{}, fmt.Errorf("slnDirPath must not be empty.")
	}

	slnDirPath, err := getSolutionPath(slnDirPath)
	if err != nil {
		return Affected{}, err
	}

	fi, err := os.Stat(slnDirPath)
	if os.IsNotExist(err) {
		return Affected{}, err
	}

	if !fi.Mode().IsDir() {
		return Affected{}, fmt.Errorf("slnDirPath must be a directory path.")
	}

	return Affected{
		slnPath:          slnDirPath,
		projectsProvider: createCliProjectsProvider(slnDirPath),
		changesProvider:  createGitChangesProvider(slnDirPath),
	}, nil
}

func (affected *Affected) GetAffectedProjects(fromSHA string, toSHA string) ([]Project, error) {
	changedFiles, err := affected.changesProvider.Get(fromSHA, toSHA)
	if err != nil {
		return []Project{}, err
	}

	allProjects, err := affected.projectsProvider.Get()
	if err != nil {
		return []Project{}, err
	}

	changedProjects := affected.filterChangedProjects(allProjects, changedFiles)

	return getAffectedProjects(changedProjects, allProjects), nil
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

		if affected[currentProject.Path] {
			continue
		}

		affected[currentProject.Path] = true

		for _, project := range all {
			if currentProject.Path == project.Path {
				continue
			}

			if contains(project.ReferencesPaths, currentProject.Path) {
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

func (affected *Affected) filterChangedProjects(allProjects []Project, changedFiles []string) []Project {
	changedProjects := []Project{}

	for _, file := range changedFiles {
		project, err := getFileProject(file, affected.slnPath, allProjects)
		if err == nil {
			changedProjects = append(changedProjects, project)
		}
	}

	return changedProjects
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
		if project.Path == name {
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

// Checks if filepath is relative to any project dir path.
// Returns project or error if not found
func getFileProject(file string, sln string, allProjects []Project) (Project, error) {
	for _, project := range allProjects {
		projectDir := filepath.Join(sln, filepath.Dir(project.Path))
		path, err := filepath.Rel(projectDir, filepath.Join(sln, file))
		if err == nil && !strings.Contains(path, "..") {
			return project, nil
		}
	}

	return Project{}, fmt.Errorf("NOT FOUND")
}
