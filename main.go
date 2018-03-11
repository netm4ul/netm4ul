package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
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
		go cmd.CreateServer(addr, &conf)

		//TODO flag enable / disable api
		addrAPI := ":" + strconv.FormatUint(uint64(config.Config.API.Port), 10)
		go cmd.CreateAPI(addrAPI, &conf)

	} else {

		ip := config.Config.Server.IP
		port := strconv.FormatUint(uint64(config.Config.Server.Port), 10)
		addr := ip + ":" + port
		go cmd.CreateClient(addr, &conf)

	}

	// handle gracefull shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("shutting down")
	os.Exit(0)
}
