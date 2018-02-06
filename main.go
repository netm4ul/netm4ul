package main

import (
	"flag"
	"strconv"

	"github.com/netm4ul/netm4ul/cmd"
	"github.com/netm4ul/netm4ul/cmd/config"
)

func init() {
	flag.BoolVar(&config.Config.IsServer, "server", false, "Set the node as server")
	flag.Parse()
}

func main() {

	conf := config.ConfigToml{}

	if config.Config.IsServer {
		// init array of nodes
		conf.Nodes = make(map[string]config.Node)

		// listen on all interface + Server port
		addr := ":" + strconv.FormatUint(uint64(config.Config.Server.Port), 10)
		cmd.CreateServer(addr, &conf)

	} else {
		ip := config.Config.Server.IP
		port := strconv.FormatUint(uint64(config.Config.Server.Port), 10)
		addr := ip + ":" + port
		cmd.CreateClient(addr, &conf)
	}
}
