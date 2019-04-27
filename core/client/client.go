package client

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/requirements"

	"github.com/netm4ul/netm4ul/core/session"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	maxRetry          = 10
	delayBetweenTries = 3 * time.Second
)

//Client is the base struct for a new node. It holds the basics informations on itself (modules...)
type Client struct {
	Session *session.Session
	// ListModule : global list of modules
	ListModule []string
	// ListModuleEnabled : global list of enabled modules
	ListModuleEnabled []string
}

// CreateClient : Connect the node to the master server
func CreateClient(s *session.Session) *Client {

	client := Client{Session: s}
	client.InitModule()
	return &client
}

//Start actually start the client and send an hello packet to the server
func (client *Client) Start() {

	var err error
	log.Info("Modules enabled :", client.ListModuleEnabled)

	for tries := 0; tries < maxRetry; tries++ {
		err = client.Connect()

		// no error, exit retry loop
		if err == nil {
			break
		}

		log.Errorf("Could not connect : %+v", err)
		log.Errorf("Retry count : %d, Max retries : %d", tries, maxRetry)
		time.Sleep(delayBetweenTries)
	}

	if err != nil {
		log.Fatal(err)
	}

	err = client.SendHello()
	if err != nil {
		log.Fatal(err)
	}

	// Recieve data
	go client.handleData()
}

func (client *Client) handleData() {

	for {
		cmd, err := client.Recv()

		// kill on socket closed.
		if err == io.EOF {
			log.Fatalf("Connection closed : %s", err.Error())
		}

		if err != nil {
			log.Error(err.Error())
			continue
		}
		// must exist, if it doesn't, the next lines shouldn't be executed
		module, exist := client.Session.Modules[cmd.Name]
		if !exist {
			continue
		}

		go func() {
			log.Debugf("Execute %s", cmd.Name)
			_, err = client.Execute(module, cmd)
			if err != nil {
				log.Error("Could not execute module : " + err.Error())
			}
		}()
	}
}

// Connect : Setup the connection to the master node
func (client *Client) Connect() error {

	var err error
	ipport := client.Session.GetServerIPPort()

	if client.Session.Config.TLSParams.UseTLS {
		client.Session.Connector.TLSConn, err = tls.Dial("tcp", ipport, client.Session.Config.TLSParams.TLSConfig)
	} else {
		client.Session.Connector.Conn, err = net.Dial("tcp", ipport)
	}

	if err != nil {
		return errors.Wrap(err, "Dialing "+ipport+" failed")
	}

	return nil
}

// InitModule : Update ListModule & ListModuleEnabled variable
func (client *Client) InitModule() {

	log.Debugf("Session client : %+v", client.Session)
	for m := range client.Session.Config.Modules {
		client.ListModule = append(client.ListModule, m)
		if client.Session.Config.Modules[m].Enabled {
			client.ListModuleEnabled = append(client.ListModuleEnabled, m)
		}
	}
}

// SendHello : Send node info (modules list, project name,...)
func (client *Client) SendHello() error {
	var rw *bufio.ReadWriter

	if client.Session.Connector.TLSConn == nil {
		rw = bufio.NewReadWriter(bufio.NewReader(client.Session.Connector.Conn), bufio.NewWriter(client.Session.Connector.Conn))
	} else {
		rw = bufio.NewReadWriter(bufio.NewReader(client.Session.Connector.TLSConn), bufio.NewWriter(client.Session.Connector.TLSConn))
	}

	compcap, err := requirements.GetCPUCapacity()
	if err != nil {
		log.Errorf("Couldn't get the CPU Stats : %s", err.Error())
	}

	memcap, err := requirements.GetMemoryCapacity()
	if err != nil {
		log.Errorf("Couldn't get the Memory Stats : %s", err.Error())
	}

	req := requirements.Requirements{
		ComputingCapacity: compcap,
		MemoryCapacity:    memcap,
		NetworkCapacity:   requirements.CapacityMedium,  // TOFIX, hardcoded network capacity
		NetworkType:       requirements.NetworkExternal, // TOFIX, hardcoded network type
	}

	// TODO : save this id into a file, might want to re-use it instead
	token := make([]byte, 5)
	rand.Read(token)
	sha := fmt.Sprintf("%x", sha1.Sum(token))
	log.Info("New node ID : ", sha)

	node := communication.Node{
		ID:           sha,
		Modules:      client.ListModuleEnabled,
		Project:      client.Session.Config.Project.Name,
		Requirements: req,
		IsAvailable:  true,
	}

	log.Debugf("Node : %+v", node)

	err = gob.NewEncoder(rw).Encode(node)
	if err != nil {
		return err
	}

	err = rw.Flush()
	if err != nil {
		return err
	}

	return nil
}

// Recv read the incomming data from the server. The server use the communication.Command struct.
func (client *Client) Recv() (communication.Command, error) {
	var cmd communication.Command

	log.Debugf("Waiting for incomming data")

	var rw *bufio.ReadWriter

	if client.Session.Connector.TLSConn == nil {
		rw = bufio.NewReadWriter(bufio.NewReader(client.Session.Connector.Conn), bufio.NewWriter(client.Session.Connector.Conn))
	} else {
		rw = bufio.NewReadWriter(bufio.NewReader(client.Session.Connector.TLSConn), bufio.NewWriter(client.Session.Connector.TLSConn))
	}

	err := gob.NewDecoder(rw).Decode(&cmd)

	// handle connection closed (server shutdown for example)
	if err == io.EOF {
		return communication.Command{}, err
	}

	if err != nil {
		return communication.Command{}, errors.New("Could not decode received message : " + err.Error())
	}

	log.Debugf("Received command %+v", cmd)
	_, ok := client.Session.Modules[cmd.Name]

	if !ok {
		return communication.Command{}, errors.New("Unsupported (or unknown) command : " + cmd.Name)
	}

	return cmd, nil
}

// Execute runs modules with the right options and handle errors.
func (client *Client) Execute(module modules.Module, cmd communication.Command) (communication.Done, error) {

	log.Debugf("Executing module : \n\t %s, version %s by %s\n\t", module.Name(), module.Version(), module.Author())
	result := make(chan communication.Result)
	done := make(chan communication.Done)

	go func() {
		for {
			select {
			case res := <-result:
				client.SendResult(res)
			case <-done:
				return
			}
		}
	}()

	// Pass the channel in every module so we can get the data back
	res, err := module.Run(cmd.Options, result)
	done <- res
	return res, err
}

// SendResult sends the data back to the server. It will then be handled by each module.WriteDb to be saved
func (client *Client) SendResult(res communication.Result) error {
	var rw *bufio.ReadWriter

	if client.Session.Connector.TLSConn == nil {
		rw = bufio.NewReadWriter(bufio.NewReader(client.Session.Connector.Conn), bufio.NewWriter(client.Session.Connector.Conn))
	} else {
		rw = bufio.NewReadWriter(bufio.NewReader(client.Session.Connector.TLSConn), bufio.NewWriter(client.Session.Connector.TLSConn))
	}

	err := gob.NewEncoder(rw).Encode(res)

	if err != nil {
		log.Errorf("Could not encode result : %+v", err)
		return err
	}

	err = rw.Flush()
	if err != nil {
		log.Errorf("Could not flush the read/writer : %+v", err)
		return err
	}

	return nil
}
