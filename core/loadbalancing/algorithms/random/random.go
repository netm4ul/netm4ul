package random

import (
	"math/rand"
	"time"

	"github.com/netm4ul/netm4ul/core/communication"
)

//Random is the struct for this algorithm
type Random struct {
	Nodes map[string]communication.Node
}

//NewRandom is a Random generator.
func NewRandom() *Random {
	rand.Seed(time.Now().UnixNano())
	r := Random{}
	return &r
}

//Name is the name of the algorithm
func (r *Random) Name() string {
	return "Random"
}

//SetNodes is the setter for the Nodes variable.
//It is used for adding new nodes from outside this package.
func (r *Random) SetNodes(nodes map[string]communication.Node) {
	r.Nodes = nodes
}

//NextExecutionNodes returns selected nodes
func (r *Random) NextExecutionNodes(cmd communication.Command) map[string]communication.Node {

	// no client found !
	if len(r.Nodes) == 0 {
		return map[string]communication.Node{}
	}
	var selectedNode communication.Node
	x := rand.Intn(len(r.Nodes))
	for _, node := range r.Nodes {
		if x == 0 {
			selectedNode = node
			break
		}
		x--
	}
	return map[string]communication.Node{selectedNode.ID: selectedNode}
}
