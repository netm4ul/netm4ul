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
	templateAlgorithm := `
package {{.algorithmName}}

import (
	"github.com/netm4ul/netm4ul/core/communication"
	log "github.com/sirupsen/logrus"
)

//{{.algorithmName}} is the struct for this algorithm
type {{.algorithmName}} struct {
	Nodes []communication.Node
}

//New{{.algorithmName}} is a {{.algorithmName}} generator.
func New{{.algorithmName}}() *{{.algorithmName}} {
	{{.algorithmShortName}} := {{.algorithmName}}{}
	return &{{.algorithmShortName}}
}

//Name is the name of the algorithm
func ({{.algorithmShortName}} *{{.algorithmName}}) Name() string {
	return "{{.algorithmName}}"
}

func ({{.algorithmShortName}} *{{.algorithmName}}) SetNodes(nodes []communication.Node) {
	{{.algorithmShortName}}.Nodes = nodes
}

//NextExecutionNodes returns selected nodes
func ({{.algorithmShortName}} *{{.algorithmName}}) NextExecutionNodes(cmd communication.Command) []communication.Node {
	selectedNode := []communication.Node{}

	return selectedNode
}
`
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

	bytes, err := scripts.GenerateSourceTemplate(templateAlgorithm, data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(filepath, bytes)
	if err != nil {
		log.Fatal(err)
	}
}
