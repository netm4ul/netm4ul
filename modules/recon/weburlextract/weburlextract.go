package WebURLExtract

import (
	"encoding/gob"
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
)

type WebURLExtractConfig struct {
}

type WebURLExtract struct {
	Config WebURLExtractConfig
}

// NewWebURLExtract generate a new WebURLExtract module (type modules.Module)
func NewWebURLExtract() modules.Module {
	gob.Register(WebURLExtract{})
	var t modules.Module
	t = &WebURLExtract{}
	return t
}

func (wue *WebURLExtract) Name() string {
	return "WebURLExtract"
}

func (wue *WebURLExtract) Version() string {
	return "1.0"
}

func (wue *WebURLExtract) Author() string {
	return Skawak
}

func (wue *WebURLExtract) DependsOn() []modules.Condition {
	return nil
}

func (wue *WebURLExtract) Run([]modules.Input) (modules.Result, error) {
	return modules.Result{}, errors.New("Not implemented yet")
}

func (wue *WebURLExtract) ParseConfig() error {
	return errors.New("Not implemented yet")
}

func (wue *WebURLExtract) WriteDb(result modules.Result, db models.Database, projectName string) error {
	return errors.New("Not implemented yet")
}
