package generate

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/netm4ul/netm4ul/scripts"
)

//GenerateAlgorithm generate boilerplate for algorithm
func GenerateAlgorithm(algorithmName, algorithmShortName string) {
	if algorithmName == "" {
		fmt.Println("You must provide an algorithm name")
		os.Exit(1)
	}

	// if no short name are provided, use the first letter of the long version, in lowercase
	if algorithmShortName == "" {
		algorithmShortName = string(strings.ToLower(algorithmName)[0])
	}

	data := map[string]string{
		"algorithmName":      algorithmName,
		"algorithmShortName": algorithmShortName,
	}

	//ensure data folder exists
	dirpath := path.Join("./core/loadbalancing/algorithms", strings.ToLower(algorithmName))
	filepath := path.Join(dirpath, strings.ToLower(algorithmName)+".go")

	bytes, err := scripts.GenerateSourceTemplate("algorithm.tmpl", "./scripts/generate/templates/algorithm.tmpl", data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.EnsureDir(filepath)
	if err != nil {
		log.Println("The directory already exist. Do you want to continue ? [y/n]")

		input := ""
		fmt.Scanln(&input)
		if input != "y" && input != "Y" {
			log.Fatal("Aborting.")
		}
	}

	err = scripts.SaveFileToPath(filepath, bytes)
	if err != nil {
		log.Fatal(err)
	}
}
