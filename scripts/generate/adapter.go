package generate

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/netm4ul/netm4ul/scripts"
)

//GenerateAdapter generate boilerplate for adapter
func GenerateAdapter(adapterName, adapterShortName string) {

	if adapterName == "" {
		fmt.Println("You must provide an adapter name")
		os.Exit(1)
	}
	// if no short name are provided, use the first letter of the long version, in lowercase
	if adapterShortName == "" {
		adapterShortName = string(strings.ToLower(adapterName)[0])
	}

	data := map[string]string{
		"adapterName":      adapterName,
		"adapterShortName": adapterShortName,
	}

	dirpath := "./core/database/adapters/" + strings.ToLower(adapterName)
	modelpath := path.Join(dirpath, "models.go")
	adapterpath := path.Join(dirpath, strings.ToLower(adapterName)+".go")

	err := scripts.EnsureDir(adapterpath)
	if err != nil {
		log.Println("The directory already exist. Do you want to continue ? [y/n]")

		input := ""
		fmt.Scanln(&input)
		if input != "y" && input != "Y" {
			log.Fatal("Aborting.")
		}
	}

	bytes, err := scripts.GenerateSourceTemplate("adapter_interface.tmpl", "./scripts/generate/templates/adapter_interface.tmpl", data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(adapterpath, bytes)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err = scripts.GenerateSourceTemplate("adapter_model.tmpl", "./scripts/generate/templates/adapter_model.tmpl", data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(modelpath, bytes)
	if err != nil {
		log.Fatal(err)
	}

}
