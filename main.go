package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/netm4ul/netm4ul/cmd"
	"github.com/netm4ul/netm4ul/cmd/cli"
	"github.com/netm4ul/netm4ul/cmd/config"
)

func init() {
	cli.ParseArgs()
}

func main() {

	if config.Config.IsServer && config.Config.IsClient {
		log.Fatalln("Cannot be Server AND Client at the same time")
	}

	// -server (master mode)
	if config.Config.IsServer {
		// init array of nodes
		config.Config.Nodes = make(map[string]config.Node)

		// listen on all interface + Server port
		addr := ":" + strconv.FormatUint(uint64(config.Config.Server.Port), 10)
		go cmd.CreateServer(addr, &config.Config)

		//TODO flag enable / disable api
		addrAPI := ":" + strconv.FormatUint(uint64(config.Config.API.Port), 10)
		go cmd.CreateAPI(addrAPI, &config.Config)

		gracefulShutdown()
		//unreachable return, only for clarity
		return
	}

	// -client (node mode)
	if config.Config.IsClient {

		ip := config.Config.Server.IP
		port := strconv.FormatUint(uint64(config.Config.Server.Port), 10)
		addr := ip + ":" + port
		go cmd.CreateClient(addr, &config.Config)

		gracefulShutdown()
		//unreachable return, only for clarity
		return
	}

	// CLI mode

}

func gracefulShutdown() {
	// handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("shutting down")
	os.Exit(0)
}
