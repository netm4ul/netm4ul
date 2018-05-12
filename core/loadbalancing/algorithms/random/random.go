package random

import (
	"math/rand"
	"time"

	"github.com/netm4ul/netm4ul/core/communication"
)

//Random is the struct for this algorithm
type Random struct {
	Nodes []communication.Node
}

//NewRandom is a Random generator.
func NewRandom() *Random {
	r := Random{}
	return &r
}

//Name is the name of the algorithm
func (r *Random) Name() string {
	return "Random"
}

func (r *Random) SetNodes(nodes []communication.Node) {
	r.Nodes = nodes
}

//NextExecutionNodes returns selected nodes
func (r *Random) NextExecutionNodes(cmd communication.Command) []communication.Node {
	rand.Seed(time.Now().UnixNano())

	selectedNode := []communication.Node{
		r.Nodes[rand.Intn(len(r.Nodes))],
	}
	return selectedNode
}
