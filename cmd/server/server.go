package server

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
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

	handleHello(rw)
}
func handleHello(rw *bufio.ReadWriter) {
	var data string

	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)

	if err != nil {
		log.Println("Cannot read gob data :", err)
		return
	}
	fmt.Println(data)
}
