package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/janhalfar/gitlog/git"
)

const docs = `
I do git log --numstat --pretty=medium --summary --date=local
    yaml as yaml
    csv as a csv
`

func usage() {
	fmt.Println("usage", os.Args[0], ":")
	fmt.Println(docs)
	os.Exit(1)

}

func dieOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func dir() string {
	if len(os.Args) > 2 {
		return os.Args[2]
	}
	return "."
}
func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "scan":
		repos := []string{}
		err := git.FindRepos(dir(), repos)
		dieOnErr(err)
		for _, r := range repos {
			fmt.Println(r)
			git.CSV(r)
		}
	case "yaml":
		log, err := git.Log(dir())
		dieOnErr(err)
		yamlBytes, err := yaml.Marshal(log)
		dieOnErr(err)
		fmt.Println(string(yamlBytes))

	case "csv":
		git.CSV(dir())
	default:
		usage()
	}

}
