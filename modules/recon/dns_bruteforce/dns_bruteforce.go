package main

import (

	"fmt"
	"os/exec"
)

func main() {
	args := []string{"@212.129.38.224", "-p", "54011", "ch11.challenge01.root-me.org", "-t", "AXFR"}
	fmt.Println("Args :", args)
	cmd := exec.Command("/usr/bin/dig", args...)
	fmt.Println("My cmd :", cmd)
	res, err := cmd.Output()
	if err != nil {
        fmt.Println("error occured")
        panic(err)
    }
	fmt.Println("\nRes :\n",string(res))
}
