package main

import (
	"fmt"
	"strconv"

	"github.com/netm4ul/netm4ul/cmd"
)

func main() {
	// TODO : Parse args
	nodetype := "server"

	fmt.Println(len(cmd.ListModuleEnabled), "enabled over", len(cmd.ListModule), "module(s) loaded")

	if nodetype == "server" {
		fmt.Println("Server mode")
		fmt.Println("[*]Listen on port :", cmd.Config.MQ.Port)
		cmd.Listen(":" + strconv.FormatUint(uint64(cmd.Config.MQ.Port), 10)) // listen on all interface + MQ port
	}

	if nodetype == "client" {
		fmt.Println("Client mode")
	}
}
