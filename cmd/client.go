package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func Connect(ipport string) {

	// connect to this socket
	conn, _ := net.Dial("tcp", ipport)
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}
