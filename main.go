package main

import (
	"flag"
	"fmt"
	"mono-sharp/pkg/affected"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {

	from := flag.String("from", "HEAD", "'from' git commit")
	to := flag.String("to", "HEAD~1", "'to' git commit")
	dir := flag.String("slnDir", "./", "solution file directory")

	flag.Parse()

	directory, err := filepath.Abs(*dir)
	if err != nil {
		panic(err)
	}

	if directory == "./" {
		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		directory = currentDir
	}

	aff, err := affected.CreateAffected(directory)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("%s", string(exitError.Stderr))
		}

		fmt.Printf("%v", err)
		os.Exit(1)
	}

	affectedProjects, err := aff.GetAffectedProjects(*from, *to)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("%s", string(exitError.Stderr))
		}

		fmt.Printf("%v", err)
		os.Exit(1)
	}

	for _, project := range affectedProjects {
		fmt.Println(project.Path)
	}
}
