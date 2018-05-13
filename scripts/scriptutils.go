package scripts

import (
	"bytes"
	"errors"
	"go/format"
	"html/template"
	"os"
	"path"
)

//GenerateSourceTemplate returns the generated template, filled with data or error.
func GenerateSourceTemplate(templateStr string, data map[string]string) ([]byte, error) {

	tmpl, err := template.New("template").Parse(templateStr)

	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
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

	dirpath := path.Dir(filepath)
	//ensure data folder exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.Mkdir(dirpath, 0755)
	} else {
		return errors.New("Folder " + dirpath + " already exist, aborting.")
	}

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		return errors.New("Could not open file " + filepath)
	}

	_, err = file.Write(data)
	return err
}
