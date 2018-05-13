package WebURLExtract

import (
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
var domain string = "https://facebook.com"

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

func (wue *WebURLExtract) Run([]modules.Input) (modules.Result, error) {

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

    fmt.Println("Domain :", domain)
    fmt.Println("Résults :")
    // fmt.Println(s)
    for i:=0;i<len(s);i++ {
        fmt.Println(" - " + s[i])
    }
    
	return modules.Result{}, errors.New("Not implemented yet")
}

func (wue *WebURLExtract) ParseConfig() error {
	return errors.New("Not implemented yet")
}

func (wue *WebURLExtract) WriteDb(result modules.Result, db models.Database, projectName string) error {
	return errors.New("Not implemented yet")
}
