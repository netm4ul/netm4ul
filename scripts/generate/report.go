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

	bytes, err := scripts.GenerateSourceTemplate("report", "./scripts/generate/templates/reports.tmpl", data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(filepath, bytes)
	if err != nil {
		log.Fatal(err)
	}
}
