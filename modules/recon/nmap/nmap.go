package nmap

//package nmap

import (
	mynmap "github.com/lair-framework/go-nmap"
	"github.com/netm4ul/netm4ul/modules"
	"io/ioutil"
	"log"
	"os/exec"
)

// Nmap "class"
type Nmap struct {
	Result  []byte
	Nmaprun *mynmap.NmapRun
}

func (N *Nmap) Name() string {
	name := "Nmap"
	return name
}

func (N *Nmap) Version() string {
	version := "v0.1"
	return version
}

func (N *Nmap) Author() string {
	author := "Author : pruno"
	return author
}

func (N *Nmap) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Running nmap
func (N *Nmap) Run(interface{}) (interface{}, error) {
	args := []string{"-A", "-v", "-oX", "nmap_127.0.0.1.xml", "-p 8080", "127.0.0.1"}
	cmd := exec.Command("/usr/bin/nmap", args...)
	execErr := cmd.Run()
	if execErr != nil {
		log.Fatal(execErr)
	}
	var err error
	N.Result, err = ioutil.ReadFile("nmap_127.0.0.1.xml")
	if err != nil {
		log.Fatal("Error 1 !", err)
	}
	return N.Result, err
}

func (N *Nmap) Parse() (interface{}, error) {
	// fmt.Println(mynmap.Parse(content))
	var err error
	N.Nmaprun, err = mynmap.Parse(N.Result)
	if err != nil {
		log.Fatal("Error 2 !", err)
	}

	return N.Nmaprun, err
}

func (N *Nmap) HandleMQ() error {
	return nil
}

func (N *Nmap) SendMQ(data []byte) error {
	return nil
}

func (N *Nmap) ParseConfig() error {
	return nil
}

// MAIN
// func main() {
// 	N := &Nmap{}
// 	// ************************************
// 	fmt.Println(strings.Repeat("#", 42))

// 	// Main info
// 	fmt.Println(N.Name())
// 	fmt.Println(N.Version())
// 	fmt.Println(N.Author())

// 	fmt.Println(strings.Repeat("#", 42))
// 	// ************************************

// 	N.Run("")
// 	data, err := N.Parse()
// 	if err != nil {
// 		log.Fatal("Error 3 !", err)
// 	}
// 	fmt.Println(data)
// }
