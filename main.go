package main

import (
	"fmt"
	"github.com/netm4ul/netm4ul/cmd"
)

func main() {
	fmt.Println(len(cmd.ListModuleEnabled), "enabled over", len(cmd.ListModule), "module(s) loaded")
	// _ = cmd.ListModule[0].ParseConfig()
}
