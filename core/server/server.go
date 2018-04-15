package server

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/netm4ul/netm4ul/modules"
	mgo "gopkg.in/mgo.v2"

	"crypto/tls"
	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server/database"
	"github.com/netm4ul/netm4ul/core/session"
)

var (
	Version = config.Config.Versions.Server
	// ConfigServer : Global config for the server. Must be goroutine safe
	ConfigServer *config.ConfigToml
	//Nodes represent a map to net.Conn
	Nodes map[string]net.Conn
	//SessionServer represent the server side's session. Hold all the modules
	SessionServer *session.Session
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

	SessionServer = session.NewSession()
}

// Listen : create the TCP server on ipport interface ("ip:port" format)
func TLSListen(ipport string, conf *config.ConfigToml) {
	log.Printf(colors.Green("Listenning on : %s"), ipport)

	var l net.Listener

	if conf.TLSParams.UseTLS {
		l, err := tls.Listen("tcp", ipport, conf.TLSParams.TLSConfig)
		if err != nil {
			log.Println(colors.Red("Error listening : %s"), err.Error())
			os.Exit(1)
		}
		defer l.Close()
	} else {
		l, err := net.Listen("tcp", ipport)
		if err != nil {
			log.Println(colors.Red("Error listening : %s"), err.Error())
			os.Exit(1)
		}
		defer l.Close()
	}

	// Close the listener when the application closes.
	mgoSession := database.Connect()

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(colors.Red("Error accepting : %s"), err.Error())
			os.Exit(1)
		}

		mgoSessionClone := mgoSession.Clone()
		// Handle connections in a new goroutine. (multi-client)
		go handleRequest(conn, mgoSessionClone)
	}
}

func handleRequest(conn net.Conn, mgoSession *mgo.Session) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	// defer conn.Close()

	handleHello(conn, rw, mgoSession)

	stop := false
	for !stop {
		stop = handleData(conn, rw, mgoSession)
	}
}

// Recv basic info for the node at connection time.
func handleHello(conn net.Conn, rw *bufio.ReadWriter, session *mgo.Session) {

	var node config.Node
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&node)

	if err != nil {
		log.Println(colors.Red("Cannot read hello data : %s"), err.Error())
		return
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	if config.Config.Verbose {
		if _, ok := ConfigServer.Nodes[ip]; ok {
			log.Println(colors.Yellow("Node known. Updating"))
		} else {
			log.Println(colors.Yellow("Unknown node. Creating"))
		}
	}

	ConfigServer.Nodes[ip] = node
	Nodes[ip] = conn
	database.CreateProject(session, node.Project)

	p := database.GetProjects(session)

	if config.Config.Verbose {
		log.Printf(colors.Yellow("Nodes : %+v"), ConfigServer.Nodes)
		log.Printf(colors.Yellow("Projects : %+v"), p)
	}

}

func getProjectByNodeIP(ip string) (string, error) {

	var err error

	n, ok := ConfigServer.Nodes[ip]
	if !ok {
		return "", errors.New("Unknown node ! Could not get project name")
	}

	project := n.Project

	return project, err
}

//SendCmdByName is a wrapper to the SendCommand function.
func SendCmdByName(name string, options []string) {
	//TODO get the Command by module name & setup options and requirements

	// cmd := Command{
	// 	Name: name,
	// 	Options: options,
	// 	Requirements: GetRequirementFromCommandName(name)
	// }
	// SendCmd(cmd)
}

//SendCmd sends one commands with its options to selected clients
func SendCmd(command Command) error {

	conns, err := getAvailableNodes(command.Requirements)

	if config.Config.Verbose {
		log.Printf(colors.Yellow("Available node(s) : %d"), len(conns))
	}
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
			return errors.New("Could not send command :" + err.Error())
		}
		log.Printf(colors.Green("Sent command \"%s\" to %s"), command.Name, conn.RemoteAddr().String())
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

// handleData decode and route all data after the "hello". It listens forever until connection closed.
func handleData(conn net.Conn, rw *bufio.ReadWriter, mgoSession *mgo.Session) bool {
	var data modules.Result
	var err error

	dec := gob.NewDecoder(rw)
	err = dec.Decode(&data)

	// handle connection closed (client shutdown)
	if err == io.EOF {
		log.Printf(colors.Red("Connection closed : %s"), err.Error())
		// stop all handleData for this conn
		return true
	}

	// handle other error
	if err != nil {
		log.Printf(colors.Red("Error while decoding data : %s"), err.Error())
		return false
	}

	module, ok := SessionServer.Modules[strings.ToLower(data.Module)]
	if !ok {
		log.Printf(colors.Red("Unknown module : %s %+v %s"), data.Module, module, ok)
	}
	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
	projectName, err := getProjectByNodeIP(ip)
	fmt.Printf("%+v", data)
	err = module.WriteDb(data, mgoSession, projectName)

	if err != nil {
		log.Println(colors.Red("Database error : %s"), err)
	}
	return false

}
