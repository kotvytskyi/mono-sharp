package main

import (
	"flag"
	"fmt"
	"mono-sharp/pkg/affected"
	"os"
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

	aff, err := affected.CreateAffected(os.Args[1])
	if err != nil {
		panic(err)
	}

	affectedProjects, err := aff.GetAffectedProjects(*from, *to)
	if err != nil {
		panic(err)
	}

	for _, project := range affectedProjects {
		fmt.Println(project.Path)
	}
}
