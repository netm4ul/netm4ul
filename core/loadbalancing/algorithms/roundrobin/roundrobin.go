package roundrobin

import (
	"github.com/netm4ul/netm4ul/core/communication"
	log "github.com/sirupsen/logrus"
)

//RoundRobin is the struct for this algorithm
type RoundRobin struct {
	Nodes map[string]communication.Node
}

//NewRoundRobin is a RoundRobin generator.
func NewRoundRobin() *RoundRobin {
	rr := RoundRobin{}
	return &rr
}

//Name is the name of the algorithm
func (rr *RoundRobin) Name() string {
	return "RoundRobin"
}

//SetNodes is the setter for the Nodes variable.
//It is used for adding new nodes from outside this package.
func (rr *RoundRobin) SetNodes(nodes map[string]communication.Node) {
	rr.Nodes = nodes
}

//NextExecutionNodes returns just 1 node every time.
func (rr *RoundRobin) NextExecutionNodes(cmd communication.Command) map[string]communication.Node {
	//Sort node by their ID
	selectedNode := map[string]communication.Node{}
	log.Debugf("rr.Nodes : %+v", rr.Nodes)
	for _, n := range rr.Nodes {
		if n.IsAvailable {
			selectedNode[n.ID] = n
			//break at the first available node
			break
		}
	}

	return selectedNode
}
