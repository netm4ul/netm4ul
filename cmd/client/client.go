package client

import (
	"bufio"
	"net"

	"github.com/pkg/errors"
)

// Connect : Setup the connection to the master node
func Connect(ipport string) (*bufio.ReadWriter, error) {
	conn, err := net.Dial("tcp", ipport)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+ipport+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}
