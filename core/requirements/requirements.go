package requirements

const (
	//CapacityLow defines the lowest tier for a performance metric
	CapacityLow Capacity = iota
	//CapacityMedium defines the middle tier for a performance metric
	CapacityMedium
	//CapacityHigh defines the highest tier for a performance metric
	CapacityHigh
)

const (
	NetworkExternal = "external"
	NetworkInternal = "internal"
)

//Capacity represent an amout of available ressources
type Capacity int

//Requirements defines all the specification needed for a node to be eligble at executing commands.
type Requirements struct {
	NetworkType       string   `json:"networkType"`       // "external", "internal", ""
	NetworkCapacity   Capacity `json:"networkCapacity"`   // CapacityLow, CapacityMedium, CapacityHigh
	ComputingCapacity Capacity `json:"computingCapacity"` // CapacityLow, CapacityMedium, CapacityHigh
}
