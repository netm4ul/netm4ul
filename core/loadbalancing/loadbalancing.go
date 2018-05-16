package loadbalancing

import (
	"errors"
	"strings"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/loadbalancing/algorithms/random"
	"github.com/netm4ul/netm4ul/core/loadbalancing/algorithms/roundrobin"
	"github.com/netm4ul/netm4ul/core/requirements"
	log "github.com/sirupsen/logrus"
)

var (
	algos    map[string]Algorithm
	usedAlgo = "Random" // TOFIX : read from config
)

//Algorithm is the interface implemented by all load balancing algorithm
type Algorithm interface {
	Name() string
	SetNodes(nodes []communication.Node)
	NextExecutionNodes(cmd communication.Command) []communication.Node
}

func init() {
	algos = make(map[string]Algorithm, 2)

	// round robin
	rr := roundrobin.NewRoundRobin()
	Register(rr)
	r := random.NewRandom()
	Register(r)
}

//NewAlgo return the selected algo from the config file
func NewAlgo() (Algorithm, error) {
	//usedAlgo is still hardfixed for now
	a, ok := algos[strings.ToLower(usedAlgo)]
	log.Debug(algos)
	if !ok {
		return nil, errors.New("Could not read the provided algorithm :" + usedAlgo)
	}

	return a, nil
}

//FilterNodes returns only the nodes meetting the requirements
func FilterNodes(nodes []communication.Node, cmd communication.Command) []communication.Node {
	reqMet := []communication.Node{}
	for _, node := range nodes {
		if testRequirement(node.Requirements, cmd.Requirements) {
			reqMet = append(reqMet, node)
		}
	}
	return reqMet
}

func testRequirement(req, reqToMet requirements.Requirements) bool {
	if req.ComputingCapacity <= reqToMet.ComputingCapacity {
		return false
	}
	if req.MemoryCapacity <= reqToMet.MemoryCapacity {
		return false
	}
	if req.NetworkCapacity <= reqToMet.NetworkCapacity {
		return false
	}
	if req.NetworkType != reqToMet.NetworkType {
		return false
	}

	return true
}

//Register a new algorithm to the algos map
func Register(algo Algorithm) {
	algos[strings.ToLower(algo.Name())] = algo
}
