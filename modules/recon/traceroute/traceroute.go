package traceroute

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/aeden/traceroute"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
)

// TracerouteResult represent the parsed ouput
type TracerouteResult struct {
	Source      string
	Destination string
	Max         float32
	Min         float32
	Avg         float32
}

// Config : configuration model (from the toml file)
type Config struct {
	MaxHops int `toml:"max_hops"`
}

// Traceroute "class"
type TracerouteModule struct {
	// Config : exported config
	Config Config
}

type Traceroute struct {
	Hops   []models.Hop
	Src    net.IP
	Dst    net.IP
	ttl    int
	maxTTL int
}

//NewTraceroute generate a new Traceroute module (type modules.Module)
func NewTraceroute() modules.Module {
	gob.Register(traceroute.TracerouteResult{})
	var t modules.Module
	t = &TracerouteModule{}
	return t
}

// Name : name getter
func (T *TracerouteModule) Name() string {
	return "Traceroute"
}

// Author : Author getter
func (T *TracerouteModule) Author() string {
	return "Edznux"
}

// Version : Version  getter
func (T *TracerouteModule) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (T *TracerouteModule) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (T *TracerouteModule) Run(inputs []modules.Input) (modules.Result, error) {

	ipAddr, err := net.ResolveIPAddr("ip", inputs[0].Domain)
	if err != nil {
		return modules.Result{}, errors.New("Could not resolve the IP : " + err.Error())
	}

	options := traceroute.TracerouteOptions{}
	options.SetMaxHops(T.Config.MaxHops)
	options.SetRetries(3)

	traceRes, err := traceroute.Traceroute(ipAddr.String(), &options)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	log.Debugf("RES : %+v\n", traceRes)

	return modules.Result{Data: traceRes, Timestamp: time.Now(), Module: T.Name()}, nil
}

// Parse : Parse the result of the execution
func (T *TracerouteModule) Parse() (TracerouteResult, error) {
	return TracerouteResult{}, nil
}

// ParseConfig : Load the config from the config folder
func (T *TracerouteModule) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "traceroute.conf")

	if _, err := toml.DecodeFile(configPath, &T.Config); err != nil {
		log.Error(err)
		return err
	}
	log.Debug(T.Config.MaxHops)
	return nil
}

// WriteDb : Save data
func (T *TracerouteModule) WriteDb(result modules.Result, db models.Database, projectName string) error {
	log.Debug("Writing to the database.")

	var data traceroute.TracerouteResult
	var err error

	data = result.Data.(traceroute.TracerouteResult)

	for _, hop := range data.Hops {

		ipDest := models.IP{
			Value:     hop.AddressString(),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}

		err = db.CreateOrUpdateIP(projectName, ipDest)
		if err != nil {
			log.Errorf("Could not create or update ip : %+v", err)
		}
	}

	err = db.AppendRawData(projectName, T.Name(), data)
	if err != nil {
		log.Errorf("Could not append : %+v", err)
	}
	return nil
}
