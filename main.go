package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/netm4ul/netm4ul/cmd"
)

func init() {
	flag.BoolVar(&cmd.Config.IsServer, "server", false, "Set the node as server")
	flag.Parse()
}

func main() {
	fmt.Println(len(cmd.ListModuleEnabled), "enabled over", len(cmd.ListModule), "module(s) loaded")

	if cmd.Config.IsServer {
		cmd.CreateServer(":" + strconv.FormatUint(uint64(cmd.Config.MQ.Port), 10)) // listen on all interface + MQ port
	} else {
		cmd.CreateClient(cmd.Config.MQ.Ip.String() + ":" + strconv.FormatUint(uint64(cmd.Config.MQ.Port), 10))
		fmt.Println("Client mode")
	}
}
