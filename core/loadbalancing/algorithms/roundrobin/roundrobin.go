package roundrobin

import (
	"github.com/netm4ul/netm4ul/core/communication"
	log "github.com/sirupsen/logrus"
)

//RoundRobin is the struct for this algorithm
type RoundRobin struct {
	Nodes []communication.Node
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

func (rr *RoundRobin) SetNodes(nodes []communication.Node) {
	rr.Nodes = nodes
}

//NextExecutionNodes returns just 1 node every time.
func (rr *RoundRobin) NextExecutionNodes(cmd communication.Command) []communication.Node {
	//Sort node by their ID
	selectedNode := []communication.Node{}
	log.Debugf("rr.Nodes : %+v", rr.Nodes)
	for _, n := range rr.Nodes {
		if n.IsAvailable {
			selectedNode = append(selectedNode, n)
			//break at the first available node
			break
		}
	}

	return selectedNode
}
