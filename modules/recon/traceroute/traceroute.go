package traceroute

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/aeden/traceroute"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
)

// Result represent the parsed output
type Result struct {
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
type Traceroute struct {
	// Config : exported config
	Config Config
}

//NewTraceroute generate a new Traceroute module (type modules.Module)
func NewTraceroute() modules.Module {
	gob.Register(traceroute.TracerouteResult{})
	var t modules.Module
	t = &Traceroute{}
	return t
}

// Name : name getter
func (T *Traceroute) Name() string {
	return "Traceroute"
}

// Author : Author getter
func (T *Traceroute) Author() string {
	return "Edznux"
}

// Version : Version  getter
func (T *Traceroute) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (T *Traceroute) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (T *Traceroute) Run(input communication.Input, resultChan chan communication.Result) (communication.Done, error) {
	var ipAddr *net.IPAddr
	var err error

	// The traceroute lib doesn't support IPV6, so we specify ipv4 only
	if input.Domain != "" {
		ipAddr, err = net.ResolveIPAddr("ip4", input.Domain)
	}
	if input.IP != nil {
		ipAddr, err = net.ResolveIPAddr("ip4", input.IP.String())
	}

	if err != nil {
		return communication.Done{Error: err}, errors.New("Could not resolve the IP : " + err.Error())
	}

	options := traceroute.TracerouteOptions{}
	options.SetMaxHops(T.Config.MaxHops)
	options.SetRetries(2)

	var traceRes traceroute.TracerouteResult

	// use channel only to print debug
	if log.GetLevel() >= log.DebugLevel {
		c := make(chan traceroute.TracerouteHop, 0)
		go func() {
			for {
				hop, ok := <-c
				if !ok {
					log.Debug("Received invalid hop (*)")
					return
				}
				log.Debugf("Received hop : %+v", hop)
			}
		}()
		traceRes, err = traceroute.Traceroute(ipAddr.String(), &options, c)
	} else {
		traceRes, err = traceroute.Traceroute(ipAddr.String(), &options)
	}

	if err != nil {
		log.Errorf("Error: %s", err)
	}

	log.Debugf("RES : %+v\n", traceRes)

	resultChan <- communication.Result{Data: traceRes, Timestamp: time.Now(), ModuleName: T.Name()}
	return communication.Done{Timestamp: time.Now(), ModuleName: T.Name()}, nil
}

// Parse : Parse the result of the execution
func (T *Traceroute) Parse() (Result, error) {
	return Result{}, nil
}

// ParseConfig : Load the config from the config folder
func (T *Traceroute) ParseConfig() error {
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
func (T *Traceroute) WriteDb(result communication.Result, db models.Database, projectName string) error {
	log.Debug("Writing to the database.")

	var data traceroute.TracerouteResult
	var err error

	data = result.Data.(traceroute.TracerouteResult)

	for _, hop := range data.Hops {

		ipDest := models.IP{
			Value:     hop.AddressString(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = db.CreateOrUpdateIP(projectName, ipDest)
		if err != nil {
			log.Errorf("Could not create or update ip : %+v", err)
		}
	}
	now := time.Now()

	dataRaws, err := json.Marshal(data)
	if err != nil {
		return err
	}

	raw := models.Raw{
		Content:    string(dataRaws),
		UpdatedAt:  now,
		CreatedAt:  now,
		ModuleName: T.Name(),
	}
	log.Debugf("raw : %+v", raw)
	log.Debugf("raw.Content : '%s'", raw.Content)
	err = db.AppendRawData(projectName, raw)

	if err != nil {
		return errors.New("Could not append raw data : " + err.Error())
	}
	return nil
}
