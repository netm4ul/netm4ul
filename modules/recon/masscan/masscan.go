package masscan

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"os"
	"path/filepath"

	"github.com/netm4ul/netm4ul/core/database/models"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/modules"
)

// MasscanResult represent the parsed ouput
type MasscanResult struct {
	Resultat []Scan
}

// Scan represents the ip and ports output
type Scan struct {
	IP        string  `json:"ip"`
	Timestamp string  `json:"timestamp"`
	Ports     []Ports `json:"ports"`
}

// Ports represents the port, proto, service, ttl, reason and status output
type Ports struct {
	Port    uint16  `json:"port"`
	Proto   string  `json:"proto"`
	Service Service `json:"service,omitempty"`
	TTL     int     `json:"ttl"`
	Reason  string  `json:"reason"`
	Status  string  `json:"status"`
}

// Service represents the name and the banner output
type Service struct {
	Name   string `json:"name"`
	Banner string `json:"banner"`
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	Verbose           bool   `toml:"verbose"`
	VeryVerbose       bool   `toml:"very-verbose"`
	Rate              int    `toml:"rate"`
	Ping              bool   `toml:"ping"`
	Seed              int    `toml:"seed"`
	Adapter           string `toml:"adapter"`
	AdapterIP         string `toml:"adapter-ip"`
	AdapterMAC        string `toml:"adapter-mac"`
	AdapterVLAN       string `toml:"adapter-vlan"`
	RouterMAC         string `toml:"router-mac"`
	Retries           int    `toml:"retries"`
	MinPacket         int    `toml:"min-packet"`
	HTTPUserAgent     string `toml:"http-user-agent"`
	RandomizeHosts    bool   `toml:"randomize-hosts"`
	Exclude           string `toml:"exclude"`
	Banners           bool   `toml:"banners"`
	Ports             string `toml:"ports"`
	ConnectionTimeout int    `toml:"connection-timeout"`
	SourcePort        int    `toml:"source-port"`
	TTL               int    `toml:"ttl"`
	Wait              string `toml:"wait"`
}

// Masscan "class"
type Masscan struct {
	// Config : exported config
	Config ConfigToml
}

//NewMasscan generate a new Masscan module (type modules.Module)
func NewMasscan() modules.Module {
	gob.Register(MasscanResult{})
	var m modules.Module
	m = &Masscan{}
	return m
}

// Name : name getter
func (M *Masscan) Name() string {
	return "Masscan"
}

// Author : Author getter
func (M *Masscan) Author() string {
	return "soldat-ryan"
}

// Version : Version  getter
func (M *Masscan) Version() string {
	return "0.3"
}

// DependsOn : Generate the dependencies requirement
func (M *Masscan) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (M *Masscan) Run(inputs []modules.Input) (modules.Result, error) {
	log.Debug("H3ll-0 M4sscan")

	// Temporary IP forced : 212.47.247.190 = edznux.fr
	outputfile := "/tmp/" + generateUUID() + ".json"

	// Get arguments
	log.Debug("Get arguments")
	opt, err := M.ParseOptions()
	if err != nil {
		log.Fatal(err)
	}
	opt = append(opt, "-oJ", outputfile)

	// Get IP
	log.Debug("Get IP")
	for _, i := range inputs {
		if i.IP != nil {
			opt = append([]string{i.IP.String()}, opt...)
		}
	}

	// Command execution
	cmd := exec.Command("masscan", opt...)
	log.Debug("Command executed: masscan ", opt)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Info(stderr.String())
	}
	log.Debug(stdout.String())
	res, err := M.Parse(outputfile)
	log.Debug("res:%+v", res)
	log.Debug("err:%+v", nil)
	if err != nil {
		log.Error(err)
	}
	log.Debug("M4sscan done.")

	return modules.Result{Data: res, Timestamp: time.Now(), Module: M.Name()}, nil
}

// ParseConfig : Load the config from the config folder
func (M *Masscan) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		log.Panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "masscan.conf")

	if _, err := toml.DecodeFile(configPath, &M.Config); err != nil {
		return err
	}
	return nil
}

// Parse : Parse the result of the execution
func (M *Masscan) Parse(file string) (MasscanResult, error) {
	var scans []Scan
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return MasscanResult{}, err
	}
	fileReformatted := string(data)

	// JSON reformatted
	re := regexp.MustCompile(" },\n]\n")
	fileReformatted = re.ReplaceAllString(string(data), " }\n]\n")

	err = json.Unmarshal([]byte(fileReformatted), &scans)
	if err != nil {
		return MasscanResult{}, err
	}

	err = os.Remove(file)
	if err != nil {
		log.Error(err)
	}
	return MasscanResult{Resultat: scans}, nil
}

// WriteDb : Save data
func (M *Masscan) WriteDb(result modules.Result, db models.Database, projectName string) error {
	log.Info("Write to the database.")

	var data MasscanResult
	data = result.Data.(MasscanResult)

	var ips []models.IP

	for _, itemIP := range data.Resultat {
		ipn := models.IP{Value: itemIP.IP}
		ips = append(ips, ipn)
		for _, itemPort := range itemIP.Ports {
			portn := models.Port{
				Number:   int16(itemPort.Port),
				Protocol: itemPort.Proto,
				Status:   itemPort.Status,
				Banner:   itemPort.Service.Banner}
			err := db.CreateOrUpdatePort(M.Name(), itemIP.IP, portn)
			if err != nil {
				log.Errorf("Could not create or update port : %+v", err)
			}
		}
	}

	err := db.CreateOrUpdateIPs(M.Name(), ips)
	if err != nil {
		log.Errorf("Could not create or update ip : %+v", err)
	}

	err = db.AppendRawData(projectName, M.Name(), data)
	if err != nil {
		log.Errorf("Could not append : %+v", err)
	}

	return nil
}

// Generate uuid name for output file
func generateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		log.Error(err)
		return "temp_masscan_output"
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// ParseOptions : Parse the args in according to masscan.conf
func (M *Masscan) ParseOptions() ([]string, error) {
	var opt []string

	err := M.ParseConfig()
	if err != nil {
		return nil, err
	}

	if M.Config.Verbose {
		opt = append(opt, "-v")
	}

	if M.Config.VeryVerbose {
		opt = append(opt, "-vv")
	}

	if M.Config.Rate != 0 {
		opt = append(opt, "--rate="+strconv.Itoa(M.Config.Rate))
	}

	if M.Config.Ping {
		opt = append(opt, "--ping")
	}

	if M.Config.Seed != 0 {
		opt = append(opt, "--seed="+strconv.Itoa(M.Config.Seed))
	}

	if M.Config.HTTPUserAgent != "" {
		opt = append(opt, "--http-user-agent="+M.Config.HTTPUserAgent)
	}

	if M.Config.Adapter != "" {
		opt = append(opt, "--adapter="+M.Config.Adapter)
	}

	if M.Config.AdapterIP != "" {
		opt = append(opt, "--adapter-ip="+M.Config.AdapterIP)
	}

	if M.Config.AdapterMAC != "" {
		opt = append(opt, "--adapter-mac="+M.Config.AdapterMAC)
	}

	if M.Config.AdapterVLAN != "" {
		opt = append(opt, "--adapter-vlan="+M.Config.AdapterVLAN)
	}

	if M.Config.RouterMAC != "" {
		opt = append(opt, "--router-mac="+M.Config.RouterMAC)
	}

	if !M.Config.RandomizeHosts {
		opt = append(opt, "--randomize-hosts="+strconv.FormatBool(M.Config.RandomizeHosts))
	}

	if M.Config.Exclude != "" {
		opt = append(opt, "--exclude="+M.Config.Exclude)
	}

	if M.Config.Banners {
		opt = append(opt, "--banners")
	}

	if M.Config.Ports == "" {
		opt = append(opt, "-p0-65535")
	} else {
		opt = append(opt, "-p"+M.Config.Ports)
	}

	if M.Config.ConnectionTimeout != 0 {
		opt = append(opt, "--connection-timeout="+strconv.Itoa(M.Config.ConnectionTimeout))
	}

	if M.Config.Retries != 0 {
		opt = append(opt, "--retries="+strconv.Itoa(M.Config.Retries))
	}

	if M.Config.MinPacket != 0 {
		opt = append(opt, "--min-packet="+strconv.Itoa(M.Config.MinPacket))
	}

	if M.Config.SourcePort != 0 {
		opt = append(opt, "--source-port="+strconv.Itoa(M.Config.SourcePort))
	}

	if M.Config.TTL != 0 {
		opt = append(opt, "--ttl="+strconv.Itoa(M.Config.TTL))
	}

	if M.Config.Wait != "" {
		opt = append(opt, "--wait="+M.Config.Wait)
	}

	return opt, nil
}
