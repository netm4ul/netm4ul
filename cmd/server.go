package cmd

import (
	"fmt"
	"io/ioutil"
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
	res, err := ioutil.ReadAll(conn)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
