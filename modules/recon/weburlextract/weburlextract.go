package WebURLExtract

import (
	
	// https://www.devdungeon.com/content/web-scraping-go
	
	"encoding/gob"
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"

	"fmt"
	"net/http"
	"strings"

    "github.com/PuerkitoBio/goquery"
)

var s[]string
var v[]string
// var domain string = "https://google.com"

// type WebURLExtractConfig struct {
// }

type WebURLExtract struct {
	Result []byte
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
	return "0.1"
}

func (wue *WebURLExtract) Author() string {
	return "Skawak"
}

func (wue *WebURLExtract) DependsOn() []modules.Condition {
	// return nil
	var _ modules.Condition
	return []modules.Condition{}
}

// This will get called for each HTML element found
func ProcessElement(index int, element *goquery.Selection) {
    // See if the href attribute exists on the element
    href, exists := element.Attr("href")
    if exists {
        // fmt.Println(href)
        if strings.HasPrefix(href, "http") {
            s = append(s, href)
        } else if strings.HasPrefix(href, "/") {
            s = append(s, domain+href)
        }
    }
}

func (wue *WebURLExtract) Run(inputs []modules.Input) (modules.Result, error) {

	log.Debug("Web URL Extract")

	var domains map[string][]string

	log.Debug("Get domains or IP")
	for _, input := range inputs {
		if input.Domain != "" {
			domains = append(domains, input.Domain)
		}
		if input.IP != nil {
			domains = append(domains, input.IP.String())
		}
	}

	log.Debug("Domains / IP recovered:", domains)

	for _, domain := range domains {
		// reset S
		s = v

		log.Debug("HTTP request to :", domain)
		// Make HTTP request
	    response, err := http.Get(domain)
	    if err != nil {
	        log.Fatal(err)
	    }
	    defer response.Body.Close()

	    // Create a goquery document from the HTTP response
	    document, err := goquery.NewDocumentFromReader(response.Body)
	    if err != nil {
	        log.Fatal("Error loading HTTP response body. ", err)
	    }

	    // Find all links and process them with the function
	    // defined earlier
	    document.Find("a").Each(ProcessElement)

	    // fmt.Println("Domain :", domain)
	    // fmt.Println("Found :", len(s), "urls")
	    // // fmt.Println(s)
	    // for i:=0;i<len(s);i++ {
	    //     fmt.Println(" - " + s[i])
	    // }

	    log.Debug("URL finds :", s)
	    domains[domain] = s
    
	}

	log.Debug("Web URL Extract done.")

	// return modules.Result{}, errors.New("Not implemented yet")
	return modules.Result{Data: domains, Timestamp: time.Now(), Module: N.Name()}, err
}

// func (wue *WebURLExtract) ParseConfig() error {
// 	return errors.New("Not implemented yet")
// }

func (wue *WebURLExtract) WriteDb(result modules.Result, db models.Database, projectName string) error {
	return errors.New("Not implemented yet")
}
