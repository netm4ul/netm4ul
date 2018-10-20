package generate

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/netm4ul/netm4ul/scripts"
)

//Module generate a new module from it's name, type and author. It implements the Module interface
func Module(name, shortName, moduleType, author string) {

	if name == "" {
		fmt.Println("You must provide an adapter name")
		os.Exit(1)
	}

	// if no short name are provided, use the first letter of the long version, in lowercase
	if shortName == "" {
		shortName = string(strings.ToLower(name)[0])
	}

	data := map[string]string{
		"name":      name,
		"shortName": shortName,
		"author":    author,
	}

	dirpath := path.Join("modules", moduleType, strings.ToLower(name))
	filepath := path.Join(dirpath, strings.ToLower(name)+".go")

	bytes, err := scripts.GenerateSourceTemplate("modules.tmpl", "./scripts/generate/templates/modules.tmpl", data)
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
