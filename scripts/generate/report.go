package generate

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/netm4ul/netm4ul/scripts"
)

//GenerateReport generate boilerplate for Report
func GenerateReport(reportName string) {
	templateReport := `
package {{.reportName | ToLower }}

import "errors"

type {{.reportName}} struct {
	Filename string
}

func Name() string {
	return "{{.reportName}}"
}

//Generate a new report in {{.reportName}} format
func (r *{{.reportName}}) Generate(name string) error {
	return errors.New("Not implemented yet")
}

`
	if reportName == "" {
		fmt.Println("You must provide a report name")
		os.Exit(1)
	}

	data := map[string]string{
		"reportName": reportName,
	}

	//ensure data folder exists
	dirpath := path.Join("./modules/report", strings.ToLower(reportName))
	filepath := path.Join(dirpath, strings.ToLower(reportName)+".go")

	bytes, err := scripts.GenerateSourceTemplate(templateReport, data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(filepath, bytes)
	if err != nil {
		log.Fatal(err)
	}
}
