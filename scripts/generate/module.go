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

	templateModule := `
package {{.name}}

import (
	"encoding/gob"
	"errors"
	log "github.com/sirupsen/logrus"
	
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
)

type {{.name}}Config struct{

}

type {{.name}} struct{
	Config {{.name}}Config
}

// New{{.name}} generate a new {{.name}} module (type modules.Module)
func New{{.name}}() modules.Module {
	gob.Register({{.name}}{})
	var t modules.Module
	t = &{{.name}}{}
	return t
}

func ({{.shortName}} *{{.name}}) Name() string{
	return "{{.name}}"
}

func ({{.shortName}} *{{.name}}) Version() string{
	return "1.0"
}

func ({{.shortName}} *{{.name}}) Author() string{
	{{ if .author }} return {{.author}}	{{ else }} return "AUTHOR NAME"	{{ end }}
}

func ({{.shortName}} *{{.name}}) DependsOn() []modules.Condition{
	return nil
}

func ({{.shortName}} *{{.name}}) Run([]modules.Input) (modules.Result, error){
	return modules.Result{}, errors.New("Not implemented yet")
}

func ({{.shortName}} *{{.name}}) ParseConfig() error{
	return errors.New("Not implemented yet")
}

func ({{.shortName}} *{{.name}}) WriteDb(result modules.Result, db models.Database, projectName string) error{
	return errors.New("Not implemented yet")
}
`
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

	bytes, err := scripts.GenerateSourceTemplate(templateModule, data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(filepath, bytes)
	if err != nil {
		log.Fatal(err)
	}
}
