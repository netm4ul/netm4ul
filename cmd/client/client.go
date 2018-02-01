package client

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/pkg/errors"
)

var (
	// ListModule : global list of modules
	ListModule []string

	// ListModuleEnabled : global list of enabled modules
	ListModuleEnabled []string
)

// Connect : Setup the connection to the master node
func Connect(ipport string) (*bufio.ReadWriter, error) {
	conn, err := net.Dial("tcp", ipport)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+ipport+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

// InitModule : Update ListModule & ListModuleEnabled variable
func InitModule() {
	for m := range config.Config.Modules {
		ListModule = append(ListModule, m)
		if config.Config.Modules[m].Enabled {
			ListModuleEnabled = append(ListModuleEnabled, m)
		}
	}
}

// SendHello : Send node info (modules list)
func SendHello(rw *bufio.ReadWriter) error {
	var err error

	enc := gob.NewEncoder(rw)

	module := ListModuleEnabled
	node := config.Node{Modules: module}

	fmt.Println(node)

	err = enc.Encode(node)
	if err != nil {
		return err
	}

	err = rw.Flush()
	if err != nil {
		return err
	}

	return nil
}
