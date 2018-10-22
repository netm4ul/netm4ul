package dnsbruteforce

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
	log "github.com/sirupsen/logrus"
)

// this wait group ensure that all the workers will stay alive for all the requests
var wg sync.WaitGroup

type dnsbruteforceConfig struct {
	WorkerCount  int    `toml:"worker_count"`
	WordlistPath string `toml:"wordlist_path"`
	Timeout      int    `toml:"timeout"` // in seconds
}
type dnsbruteforce struct {
	Config dnsbruteforceConfig
}

//Result hold the results
type Result struct {
	Domain string
	Addr   string
}

// NewDnsbruteforce generate a new dnsbruteforce module (type modules.Module)
func NewDnsbruteforce() modules.Module {
	gob.Register(Result{})
	var t modules.Module
	t = &dnsbruteforce{}
	return t
}

//Name returns the module name
func (d *dnsbruteforce) Name() string {
	return "dnsbruteforce"
}

//Version returns the module version
func (d *dnsbruteforce) Version() string {
	return "0.1"
}

//Author returns the module author
func (d *dnsbruteforce) Author() string {
	return "Edznux"
}

//DependsOn returns the module dependencies
func (d *dnsbruteforce) DependsOn() []modules.Condition {
	return nil
}

// try to generate a non existing sub domain. If it resolve, the DNS is probably resolving to all requests.
func (d *dnsbruteforce) checkWildcard(domain string) bool {
	random := time.Now().UnixNano()
	testDomain := fmt.Sprintf("%d.%s", random, domain)

	_, err := net.LookupHost(testDomain)

	// if the domain resolve (no error), then it is probably a wildcard
	if err == nil {
		log.Printf("The domain seems to have wildcard lookup : %s resolved", testDomain)
		return true
	}

	return false
}

func (d *dnsbruteforce) worker(inputList chan string, outputList chan Result) {
	for {
		testDomain := <-inputList
		log.Debugf("Got sub domain : %s", testDomain)

		ips, err := net.LookupHost(testDomain)
		if err != nil {
			//not found, just try the next one
			wg.Done()
			continue
		}

		// send all the ip found
		for _, ip := range ips {
			outputList <- Result{Addr: ip, Domain: testDomain}
		}
		wg.Done()
	}
}

//Run is the "main" function of the modules.
func (d *dnsbruteforce) Run(input communication.Input, resultChan chan communication.Result) (communication.Done, error) {
	err := d.ParseConfig()
	if err != nil {
		err := errors.New("Could not parse the config file")
		return communication.Done{Error: err}, err
	}

	log.Debugf("Starting dns bruteforcing. [Worker count : %d, Wordlist : %s, Timeout : %d]", d.Config.WorkerCount, d.Config.WordlistPath, d.Config.Timeout)
	if input.Domain == "" {
		err := errors.New("No domain name provided")
		return communication.Done{Error: err}, err
	}

	isWildcard := d.checkWildcard(input.Domain)
	if isWildcard {
		err := errors.New("The domain is a wildcard domain")
		return communication.Done{Error: err}, err
	}

	// launch all workers
	inputList := make(chan string)
	outputList := make(chan Result)

	worker := 0
	for worker < d.Config.WorkerCount {
		log.Debugf("Starting worker : %d/%d)", worker, d.Config.WorkerCount)
		go d.worker(inputList, outputList)
		worker++
	}

	// print all the domains found
	go func() {
		log.Debug("Started listening from workers")
		for {
			select {
			case found := <-outputList:
				log.Debug("Found domain : " + found.Domain + " at " + found.Addr)
				res := Result{Addr: found.Addr, Domain: found.Domain}
				resultChan <- communication.Result{Data: res, Timestamp: time.Now(), ModuleName: d.Name()}
			}

		}
	}()

	// read the wordlist and pass every line to the inputList chan
	// exit when it's done
	file, err := os.Open("wordlists/domains.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		inputList <- line + "." + input.Domain // prepend the domain with the subdomain name
		wg.Add(1)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	wg.Wait()

	return communication.Done{ModuleName: d.Name(), Timestamp: time.Now()}, nil
}

//ParseConfig load and parse the module config file
func (d *dnsbruteforce) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "dns-bruteforce.conf")
	_, err = toml.DecodeFile(configPath, &d.Config)

	if err != nil {
		log.Errorf("Couldn't parse dns-bruteforce's config file : " + err.Error())
		return err
	}

	return nil
}

//WriteDb save the result in the database
func (d *dnsbruteforce) WriteDb(result communication.Result, db models.Database, projectName string) error {

	res := result.Data.(Result)

	raw := models.Raw{Content: res.Domain + ":" + res.Addr, CreatedAt: time.Now(), UpdatedAt: time.Now(), ModuleName: d.Name()}
	err := db.AppendRawData(projectName, raw)
	if err != nil {
		return err
	}

	domain := models.Domain{Name: res.Domain, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = db.CreateOrUpdateDomain(projectName, domain)
	if err != nil {
		return err
	}

	ip := models.IP{Value: res.Addr, CreatedAt: time.Now(), UpdatedAt: time.Now(), Network: "external"}
	err = db.CreateOrUpdateIP(projectName, ip)
	if err != nil {
		return err
	}

	return nil
}
