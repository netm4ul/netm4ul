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
	templateAdapter := `
package {{.adapterName}}

import(
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
)

type {{.adapterName}} struct {
	cfg *config.ConfigToml
}

// General purpose functions
func ({{.adapterShortName}} *{{.adapterName}}) Name() string{
	return "{{.adapterName}}"
}

func ({{.adapterShortName}} *{{.adapterName}}) SetupAuth(username, password, dbname string) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) Connect(*config.ConfigToml) error{
	return errors.New("Not implemented yet")
}

// Project
func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdateProject(projectName string) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetProjects() ([]models.Project, error){
	return []models.Project{}, errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetProject(projectName string) (models.Project, error){
	return models.Project{}, errors.New("Not implemented yet")
}

// IP
func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdateIP(projectName string, ip models.IP) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdateIPs(projectName string, ip []models.IP) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetIPs(projectName string) ([]models.IP, error){
	return []models.IP{}, errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetIP(projectName string, ip string) (models.IP, error){
	return models.IP{}, errors.New("Not implemented yet")
}

// Port
func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdatePort(projectName string, ip string, port models.Port) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdatePorts(projectName string, ip string, port []models.Port) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetPorts(projectName string, ip string) ([]models.Port, error){
	return []models.Port{}, errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetPort(projectName string, ip string, port string) (models.Port, error){
	return models.Port{}, errors.New("Not implemented yet")
}

// URI (directory and files)
func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetURIs(projectName string, ip string) ([]models.URI, error){
	return []models.URI{},errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetURI(projectName string, ip string, port string, uri string) (models.URI, error){
	return models.URI{}, errors.New("Not implemented yet")
}

// Raw data
func ({{.adapterShortName}} *{{.adapterName}}) AppendRawData(projectName string, moduleName string, data interface{}) error{
	return errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetRaws(projectName string) (models.Raws, error){
	var raws models.Raws
	return raws, errors.New("Not implemented yet")
}

func ({{.adapterShortName}} *{{.adapterName}}) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	return nil, errors.New("Not implemented yet")
}
`

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
	filepath := path.Join(dirpath, strings.ToLower(adapterName)+".go")

	bytes, err := scripts.GenerateSourceTemplate(templateAdapter, data)
	if err != nil {
		log.Fatal(err)
	}

	err = scripts.SaveFileToPath(filepath, bytes)
	if err != nil {
		log.Fatal(err)
	}

}
