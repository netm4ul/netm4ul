package dnsbruteforce

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
	log "github.com/sirupsen/logrus"
)

type dnsbruteforceConfig struct {
	WorkerCount  int    `toml:"worker_count"`
	WordlistPath string `toml:"wordlist_path"`
	Timeout      int    `toml:"timeout"` // in seconds
}
type dnsbruteforce struct {
	Config dnsbruteforceConfig
}

type result struct {
	Domain string
	Addr   string
}

// NewDnsbruteforce generate a new dnsbruteforce module (type modules.Module)
func NewDnsbruteforce() modules.Module {
	gob.Register(dnsbruteforce{})
	var t modules.Module
	t = &dnsbruteforce{}
	return t
}

func (d *dnsbruteforce) Name() string {
	return "dnsbruteforce"
}

func (d *dnsbruteforce) Version() string {
	return "0.1"
}

func (d *dnsbruteforce) Author() string {
	return "Edznux"
}

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

func (d *dnsbruteforce) worker(inputList chan string, outputList chan result, stop chan struct{}) {
	log.Debug("Worker started")
	for {
		select {
		case testDomain := <-inputList:
			log.Debugf("Got sub domain : %s", testDomain)

			ips, err := net.LookupHost(testDomain)
			if err != nil {
				//not found, just try the next one
				continue
			}

			// send all the ip found
			for _, ip := range ips {
				outputList <- result{Addr: ip, Domain: testDomain}
			}
		case <-stop:
			break
		}
	}
}

func (d *dnsbruteforce) Run(input modules.Input) (modules.Result, error) {
	res := modules.Result{}

	if input.Domain == "" {
		return modules.Result{}, errors.New("No domain name provided")
	}

	isWildcard := d.checkWildcard(input.Domain)
	if isWildcard {
		return modules.Result{}, errors.New("The domain is a wildcard domain")
	}

	// launch all workers
	inputList := make(chan string)
	outputList := make(chan result)
	stop := make(chan struct{})

	worker := 0
	for worker < d.Config.WorkerCount {
		go d.worker(inputList, outputList, stop)
	}

	// print all the domains found
	go func() {
		for {
			select {
			case found := <-outputList:
				log.Println("Found domain : " + found.Domain + " at " + found.Addr)
			case <-stop:
				break
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
		inputList <- line
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return res, nil
}

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

func (d *dnsbruteforce) WriteDb(result modules.Result, db models.Database, projectName string) error {
	return errors.New("Not implemented yet")
}
