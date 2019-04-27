package certificatetransparency

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/events"
	"github.com/netm4ul/netm4ul/modules"
)

type certificateTransparencyConfig struct {
}

type certificateTransparency struct {
	Config certificateTransparencyConfig
}
type crtshResponse struct {
	IssuerCaId int    `json:"issuer_ca_id"`
	IssuerName string `json:"issuer_name"`
	NameValue  string `json:"name_value"`
	MinCertId  int    `json:"min_cert_id"`
	// MinEntryTimestamp time.Time `json:"min_entry_timestamp"`
	// NotBefore         time.Time `json:"not_before"`
	// NotAfter          time.Time `json:"not_after"`
}

// Newcertificatetransparency generate a new certificatetransparency module (type modules.Module)
func Newcertificatetransparency() modules.Module {
	gob.Register(certificateTransparency{})
	var t modules.Module
	t = &certificateTransparency{}
	return t
}

//Name returns the module name
func (ct *certificateTransparency) Name() string {
	return "certificatetransparency"
}

//Version returns the module version
func (ct *certificateTransparency) Version() string {
	return "1.0"
}

//Author returns the module author
func (ct *certificateTransparency) Author() string {
	return "Edznux"
}

//DependsOn returns the module dependencies
func (ct *certificateTransparency) DependsOn() events.EventType {
	return events.EventDomain
}

//Run is the "main" function of the modules.
func (ct *certificateTransparency) Run(input communication.Input, resultChan chan communication.Result) (communication.Done, error) {
	log.Debugln("Run certificate transparency search")

	if len(strings.Split(input.Domain.Name, ".")) > 2 {
		return communication.Done{}, errors.New("Not going to check child domain for wildcard (already included in the first domain.TLD")
	}

	var domainsFound map[string]bool // we want only unique domains
	var domainsFoundList []string    // convert to real array for the rest of the app
	domainsFound = make(map[string]bool)
	var data []crtshResponse

	crtshURL := "https://crt.sh/"

	req, err := http.NewRequest("GET", crtshURL, nil)
	if err != nil {
		log.Print(err)
		return communication.Done{}, fmt.Errorf("Could not create get request for crt.sh : %s", err)
	}

	q := req.URL.Query()
	q.Add("Identity", "%"+input.Domain.Name)
	q.Add("output", "json")
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	log.Debug(req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		return communication.Done{}, fmt.Errorf("Could not get reponse from request to crt.sh : %s", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return communication.Done{}, fmt.Errorf("Could not read body from request to crt.sh : %s", err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return communication.Done{}, fmt.Errorf("Could not parse JSON freom request to crt.sh : %s", err)
	}
	// this is crap. (double for loop)
	// It search through all inputs, put in a map (set) then convert it back to an array (unique, this time.)
	for i := range data {
		domainsFound[data[i].NameValue] = true
	}
	for d := range domainsFound {
		domainsFoundList = append(domainsFoundList, d)
	}

	log.Infof("Found domains : %+v", domainsFoundList)
	resultChan <- communication.Result{Data: domainsFoundList, Timestamp: time.Now(), ModuleName: ct.Name()}

	return communication.Done{}, nil
}

//ParseConfig load and parse the module config file
func (ct *certificateTransparency) ParseConfig() error {
	return nil
}

//WriteDb save the result in the database
func (ct *certificateTransparency) WriteDb(result communication.Result, db models.Database, projectName string) error {
	log.Debug("Write raw results to the database.")
	for _, domain := range result.Data.([]string) {
		element := models.Domain{Name: domain, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		log.Debugf("Saving IP address : %+v", element)
		err := db.CreateOrUpdateDomain(projectName, element)
		if err != nil {
			return errors.New("Could not save the database : " + err.Error())
		}
	}
	return nil
}
