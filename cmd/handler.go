package cmd

import (
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	"github.com/netm4ul/netm4ul/cmd/client"
	"github.com/netm4ul/netm4ul/cmd/server"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/recon/traceroute"
)

// ListModule : global getter for the list of modules
var ListModule []modules.Module

// ListModuleEnabled : global getter for the list of enabled modules
var ListModuleEnabled []modules.Module

/*
Initialise all the modules
always called (even in non-main package) & before main function
New modules must be included here
*/
func init() {
	var t modules.Module = &traceroute.Traceroute{}
	ListModule = append(ListModule, t)
	fmt.Println("[*] Modules loaded :")

	for _, m := range ListModule {
		if Config.Modules[strings.ToLower(m.Name())].Enabled {
			fmt.Println("\t [+]", m.Name(), "Version :", m.Version(), "Enabled !")
			ListModuleEnabled = append(ListModuleEnabled, m)
			err := m.ParseConfig()
			if err != nil {
				fmt.Println("Error: could not parse config (PerseConfig)")
				panic(err)
			}
		} else {
			fmt.Println("\t [-]", m.Name(), "Version :", m.Version(), "Disabled !")
		}
	}
}

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(ipport string) {
	server.Listen(ipport)
}

// CreateClient : Connect the node to the master server
func CreateClient(ipport string) {
	rw, err := client.Connect(ipport)
	if err != nil {
		log.Fatal(err)
	}
	enc := gob.NewEncoder(rw)
	err = enc.Encode("test")
	if err != nil {
		log.Fatal(err)
	}
	err = rw.Flush()
}
