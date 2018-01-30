package cmd

import (
	"fmt"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/recon/traceroute"
	"strings"
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
			if err != nil{
				fmt.Println("Error: could not parse config (PerseConfig)")
				panic(err)
			}
		} else {
			fmt.Println("\t [-]", m.Name(), "Version :", m.Version(), "Disabled !")
		}
	}
}
