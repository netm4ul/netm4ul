package requirements

import (
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
	NetworkType       NetworkType `json:"networkType"`       // "external", "internal", ""
	NetworkCapacity   Capacity    `json:"networkCapacity"`   // CapacityLow, CapacityMedium, CapacityHigh
	ComputingCapacity Capacity    `json:"computingCapacity"` // CapacityLow, CapacityMedium, CapacityHigh
	MemoryCapacity    Capacity    `json:"memoryCapacity"`    // CapacityLow, CapacityMedium, CapacityHigh
}

const (
	//CapacityLow defines the lowest tier for a performance metric
	CapacityLow Capacity = iota
	//CapacityMedium defines the middle tier for a performance metric
	CapacityMedium
	//CapacityHigh defines the highest tier for a performance metric
	CapacityHigh
)

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
	NetworkExternal NetworkType = iota
	NetworkInternal
)

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
	KiloByte = (1 << 10)
	MegaByte = (KiloByte << 10)
	GigaByte = (MegaByte << 10)
	TeraByte = (GigaByte << 10)
)

var (
	MemUsage map[Capacity]int64
	CPUUsage map[Capacity]int64
	NetUsage map[Capacity]int64

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
