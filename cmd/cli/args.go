package cli

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/netm4ul/netm4ul/cmd/api"
	"github.com/netm4ul/netm4ul/cmd/client"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server"
)

const (
	DefaultConfigPath = "netm4ul.conf"
	DefaultMode       = "stealth"
)

var (
	//ConfigPath reprensent the config path for the current instance
	ConfigPath string
	//Targets is a comma separated list string of targets
	Targets string
	//Modules is a comma separated list string of modules
	Modules string
	//Mode is the execution mode
	Mode string
	//Verbose enable debugs output
	Verbose bool
	version bool

	isServer bool
	isClient bool
)

// ParseArgs Parse CLI arguments
func ParseArgs() {

	// CLI commands
	flag.StringVar(&ConfigPath, "config", DefaultConfigPath, "Custom config file path")
	flag.StringVar(&Targets, "targets", "", "List of targets, comma separated")
	flag.StringVar(&Mode, "mode", DefaultMode, "Mode of execution. Simple alias to list of module. See the config file")
	flag.StringVar(&Modules, "modules", "", "List of modules executed")
	flag.BoolVar(&Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&version, "version", false, "Print the version")

	// Node setup
	flag.BoolVar(&isServer, "server", false, "Set the node as server")
	flag.BoolVar(&isClient, "client", false, "Set the node as client")

	flag.Parse()

	if version {
		PrintVersion()
	}

	config.LoadConfig(ConfigPath)

	config.Config.IsServer = isServer
	config.Config.IsClient = isClient

	if Targets == "" {
		fmt.Println("You must provide a target with '-targets <ip|domain>'")
		os.Exit(1)
	}

	ts, err := parseTargets(Targets)
	if err != nil {
		fmt.Println("Could not read -targets arguments :" + err.Error())
		os.Exit(1)
	}
	config.Config.Targets = ts

	if Modules != "" {
		mods, err := parseModules(Modules)
		if err != nil {
			fmt.Println("WARNING : " + err.Error())
		}
		addModules(mods)
	}
}

func parseTargets(str string) ([]string, error) {

	var res []string
	splitted := strings.Split(str, ",")

	if len(splitted) == 0 {
		return nil, errors.New("Not target found")
	}

	// loop on each targets
	for _, target := range splitted {
		ip, ipNet, err := net.ParseCIDR(target)

		// if this is a domain
		if err != nil {
			ips, err := net.LookupIP(target)

			if err != nil {
				return nil, err
			}

			if ips == nil {
				return nil, errors.New("Could not resolve :" + target)
			}

			// convert ips to strings
			for _, i := range ips {
				res = append(res, i.String())
			}
		} else {
			// if this is an ip

			// check if ip is specified (not :: or 0.0.0.0)
			if ip.IsUnspecified() {
				return nil, errors.New("Target ip is Unspecified (0.0.0.0 or ::)")
			}

			// check if ip is specified (not :: or 0.0.0.0)
			if ip.IsLoopback() {
				return nil, errors.New("Target ip is loopback address")
			}

			// IP Range (CIDR)
			if ipNet != nil {
				h, err := hosts(target)
				if err != nil {
					return nil, errors.New("Target ip range is invalid (" + err.Error() + ")")
				}
				res = append(res, h...)
			}
		}
	}

	return res, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func parseModules(str string) ([]string, error) {
	splitted := strings.Split(str, ",")

	if len(splitted) == 0 {
		return nil, errors.New("Could not parse modules")
	}

	for _, name := range splitted {
		_, ok := config.Config.Modules[name]
		if !ok {
			return nil, errors.New("Could not find module : " + name)
		}
	}

	return splitted, nil
}

func addModules(mods []string) {

	found := false
	for _, mod := range mods {
		for cmodname, cmod := range config.Config.Modules {
			if cmodname == mod && cmod.Enabled {
				found = true
			}
		}
		if !found {
			config.Config.Modules[mod] = config.Module{Enabled: true}
		}
	}
}

//PrintVersion Prints the version of all the components : The server, the Client, and the HTTP API
func PrintVersion() {
	fmt.Printf("Version :\n - Server : %s\n - Client : %s\n - HTTP API : %s\n", server.Version, client.Version, api.Version)
	os.Exit(0)
}
