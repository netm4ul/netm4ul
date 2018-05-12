package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
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

	tmpl, err := template.New("algorithm").Parse(templateAlgorithm)

	if err != nil {
		panic(err)
	}

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
	algorithmDirPath := path.Join("./core/loadbalancing/algorithms", strings.ToLower(algorithmName))
	if _, err := os.Stat(algorithmDirPath); os.IsNotExist(err) {
		os.Mkdir(algorithmDirPath, 0755)
	} else {
		log.Fatalf("Folder %s already exist, aborting.", algorithmDirPath)
	}
	algorithmFilePath := path.Join(algorithmDirPath, strings.ToLower(algorithmName)+".go")
	algorithmFile, err := os.OpenFile(algorithmFilePath, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		log.Fatalf("Could not open file %s", algorithmFilePath)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	p, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	algorithmFile.Write(p)
}
