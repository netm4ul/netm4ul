package server

import (
	"bufio"
	"encoding/gob"
	"errors"
	"log"
	"net"
	"os"
	"strings"

	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server/database"
)

var (
	// ConfigServer : Global config for the server. Must be goroutine safe
	ConfigServer *config.ConfigToml
	//Nodes represent a map to net.Conn
	Nodes map[string]net.Conn
)

const (
	//CapacityLow defines the lowest tier for a performance metric
	CapacityLow = 1
	//CapacityMedium defines the middle tier for a performance metric
	CapacityMedium = 2
	//CapacityHigh defines the highest tier for a performance metric
	CapacityHigh = 3
)

//Requirements defines all the specification needed for a node to be eligble at executing on command.
type Requirements struct {
	NetworkType        string `json:"networktype"`        // "external", "internal", ""
	ConnectionCapacity uint16 `json:"connectioncapacity"` // CapacityLow, CapacityMedium, CapacityHigh
	ComputingCapacity  uint16 `json:"computingcapacity"`  // CapacityLow, CapacityMedium, CapacityHigh
}

//Command represents the communication protocol between clients and the master node
type Command struct {
	Name         string       `json:"name"`
	Options      []string     `json:"options"`
	Requirements Requirements `json:"requirements"`
}

func init() {
	Nodes = make(map[string]net.Conn)
}

// Listen : create the TCP server on ipport interface ("ip:port" format)
func Listen(ipport string) {
	log.Println("Listenning : ", ipport)
	l, err := net.Listen("tcp", ipport)

	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on " + ipport)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine. (multi-client)
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	// defer conn.Close()

	handleHello(conn, rw)
}

// Recv basic info for the node at connection time.
func handleHello(conn net.Conn, rw *bufio.ReadWriter) {

	var node config.Node
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&node)

	if err != nil {
		log.Println("Cannot read hello data :", err)
		return
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	if _, ok := ConfigServer.Nodes[ip]; ok {
		log.Println("Node known. Updating")
	} else {
		log.Println("unknown node. Creating")
	}

	ConfigServer.Nodes[ip] = node
	Nodes[ip] = conn

	session := database.Connect()
	database.CreateProject(session, node.Project)
	log.Println(ConfigServer.Nodes)
	p := database.GetProjects(session)
	log.Println(p)

}

//SendCmdByName is a wrapper to the SendCommand function.
func SendCmdByName(name string, option []string) {
	//TODO get the Command by module name & setup options and requirements

	// cmd := Command{
	// 	Name: name,
	// 	Options: option,
	// 	Requirements: GetRequirementFromCommandName(name)
	// }
	// SendCmd(cmd)
}

//SendCmd sends one commands with its options to selected clients
func SendCmd(command Command) error {

	conns, err := getAvailableNodes(command.Requirements)
	log.Println("Nodes available : ", len(conns))
	if err != nil {
		return errors.New("Could not get nodes :" + err.Error())
	}

	// Send to all nodes following the requirements
	for _, conn := range conns {

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		err := gob.NewEncoder(rw).Encode(command)

		if err != nil {
			return errors.New("Could not send command :" + err.Error())
		}

		err = rw.Flush()
		if err != nil {
			log.Println(err)
			return errors.New("Could not send command :" + err.Error())
		}
		log.Println("Sent command : ", command)
	}

	return nil
}

//getAvailableNodes return a list of net.Conn available. They must follows the requirements.
func getAvailableNodes(req Requirements) ([]net.Conn, error) {
	// TODO : Requirements for each modules and load balance
	var availables []net.Conn
	for _, conn := range Nodes {
		availables = append(availables, conn)
	}
	return availables, nil
}
