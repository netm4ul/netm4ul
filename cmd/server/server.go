package server

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/netm4ul/netm4ul/cmd/config"
)

var (
	// ConfigServer : Global config for the server. Must be goroutine safe
	ConfigServer *config.ConfigToml
)

// Listen : create the TCP server on ipport interface ("ip:port" format)
func Listen(ipport string) {

	l, err := net.Listen("tcp", ipport)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + ipport)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine. (multi-client)
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	handleHello(conn, rw)
}

// Recv basic info for the node at connection time.
func handleHello(conn net.Conn, rw *bufio.ReadWriter) {

	var data config.Node
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)

	if err != nil {
		log.Println("Cannot read hello data :", err)
		return
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	if _, ok := ConfigServer.Nodes[ip]; ok {
		fmt.Println("Node known. Updating")
	} else {
		fmt.Println("unknown node. Creating")
	}

	ConfigServer.Nodes[ip] = data

	fmt.Println(ConfigServer.Nodes)
}
