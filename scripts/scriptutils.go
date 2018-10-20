package scripts

import (
	"bytes"
	"errors"
	"go/format"
	"os"
	"path"
	"strings"
	"text/template"
)

//GenerateSourceTemplate returns the generated template, filled with data or error.
func GenerateSourceTemplate(name string, templatePath string, data map[string]string) ([]byte, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	err = tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return nil, err
	}

	p, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return p, nil
}

//SaveFileToPath will try to save the file in the filepath provided.
//It will create the directory, and return an error if it already exist
func SaveFileToPath(filepath string, data []byte) error {

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return errors.New("Could not open file " + filepath)
	}

	_, err = file.Write(data)
	return err
}

//EnsureDir returns an error if the folder already exist. This is wanted in the case we don't want to override it.
func EnsureDir(filepath string) error {
	dirpath := path.Dir(filepath)
	//ensure data folder exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.Mkdir(dirpath, 0755)
		return nil
	}
	return errors.New("Folder " + dirpath + " already exist, aborting.")
}
