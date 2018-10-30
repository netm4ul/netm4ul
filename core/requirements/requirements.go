package requirements

import (
	"encoding/json"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
)

//Capacity represents the amout of available ressources
type Capacity int

//NetworkType represents the network type of the system : internal or external
type NetworkType int

//Requirements defines all the specification needed for a node to be eligble at executing commands.
type Requirements struct {
	NetworkType       NetworkType `json:"network_type"`       // "external", "internal", ""
	NetworkCapacity   Capacity    `json:"network_capacity"`   // CapacityLow, CapacityMedium, CapacityHigh
	ComputingCapacity Capacity    `json:"computing_capacity"` // CapacityLow, CapacityMedium, CapacityHigh
	MemoryCapacity    Capacity    `json:"memory_capacity"`    // CapacityLow, CapacityMedium, CapacityHigh
}

func (r Requirements) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	m["network_type"] = r.NetworkType.String()
	m["network_capacity"] = r.NetworkCapacity.String()
	m["computing_capacity"] = r.ComputingCapacity.String()
	m["memory_capacity"] = r.MemoryCapacity.String()
	return json.Marshal(m)
}

const (
	//CapacityLow defines the lowest tier for a performance metric
	CapacityLow Capacity = iota
	//CapacityMedium defines the middle tier for a performance metric
	CapacityMedium
	//CapacityHigh defines the highest tier for a performance metric
	CapacityHigh
)

//String will expand the capacity to it's human readable value
func (c Capacity) String() string {
	switch c {
	case CapacityLow:
		return "Capacity low"
	case CapacityMedium:
		return "Capacity medium"
	case CapacityHigh:
		return "Capacity high"
	default:
		return "Unknown capacity"
	}
}

const (
	//NetworkExternal represent a node with outside network access
	NetworkExternal NetworkType = iota
	//NetworkInternal represent a node with insider network access (on the target LAN)
	NetworkInternal
)

//String will expand the network type to it's human readable value
func (t NetworkType) String() string {
	switch t {
	case NetworkExternal:
		return "External Network"
	case NetworkInternal:
		return "Internal Network"
	default:
		return "Unknown type of network"
	}
}

const (
	// KiloByte represent 1024 bytes
	KiloByte = (1 << 10)
	// MegaByte represent 1024 kilobytes
	MegaByte = (KiloByte << 10)
	// GigaByte represent 1024 megabytes
	GigaByte = (MegaByte << 10)
	// TeraByte represent 1024 gigabytes
	TeraByte = (GigaByte << 10)
)

var (
	//MemUsage represents the memory capacity of the node
	MemUsage map[Capacity]int64
	//CPUUsage represents the CPU capacity of the node
	CPUUsage map[Capacity]int64
	//NetUsage represents the networking capacity of the node
	NetUsage map[Capacity]int64

	//Usage represent all capacities of the host node
	Usage = map[string]map[Capacity]int64{
		"memory":  MemUsage,
		"cpu":     CPUUsage,
		"network": NetUsage,
	}
)

func init() {
	MemUsage = map[Capacity]int64{
		CapacityLow:    1 * GigaByte,
		CapacityMedium: 2 * GigaByte,
		CapacityHigh:   4 * GigaByte,
	}

	CPUUsage = map[Capacity]int64{
		CapacityLow:    1,
		CapacityMedium: 2,
		CapacityHigh:   4,
	}
}

//GetMemoryCapacity returns a Capacty for the system memory
func GetMemoryCapacity() (Capacity, error) {
	//set cap to the lowest tier
	cap := CapacityLow

	m, err := mem.VirtualMemory()
	//if something fail, fall back to the lowest tier
	if err != nil {
		return cap, err
	}

	cap, err = returnCapacityForValue(int64(m.Available), "memory")
	if err != nil {
		return cap, err
	}

	log.Infof("Memory capacity : %s", cap)
	return cap, nil
}

//GetCPUCapacity returns the corresponding CPU Capacity
func GetCPUCapacity() (Capacity, error) {
	count, err := cpu.Counts(true)

	if err != nil {
		return CapacityLow, err
	}

	cap, err := returnCapacityForValue(int64(count), "cpu")
	if err != nil {
		return cap, err
	}

	log.Infof("CPU Capacity : %s", cap)
	return cap, nil
}

func returnCapacityForValue(value int64, capType string) (Capacity, error) {
	cap := CapacityLow
	//  >= high
	if value >= Usage[capType][CapacityHigh] {
		return CapacityHigh, nil
	}
	// < high, >= medium
	if value >= Usage[capType][CapacityMedium] {
		return CapacityMedium, nil
	}
	// < Medium
	if value <= Usage[capType][CapacityMedium] {
		return CapacityLow, nil
	}
	return cap, nil
}
