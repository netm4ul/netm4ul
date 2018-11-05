package events

import (
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
	"time"
)

//EventType represents a new type events in the application
type EventType int

func (ev EventType) String() string {
	switch ev {
	case EventIP:
		return "Event IP"
	case EventDomain:
		return "Event Domain"
	case EventPort:
		return "Event Port"
	case EventURI:
		return "Event URI"
	default:
		return "Unknown event"
	}
}

//Event is the base struct for every events. The data will be dependant of the Type attribute.
//The data should be the model (if applicable) of the event. (IP => models.IP)
type Event struct {
	Type      EventType
	Data      interface{}
	Timestamp time.Time
}

const (
	//EventIP : new IP found
	EventIP EventType = iota
	//EventDomain : new Domain found
	EventDomain
	//EventPort : new Port found
	EventPort
	//EventURI : new URI found
	EventURI
)

var (
	//EventQueue represent the global event queue for the application. Every new events will be sent using this chan
	EventQueue chan Event
)

func init() {
	// Create the event queue before anything else
	EventQueue = make(chan Event)
}

//NewEventIP send new pre-filled event of the IP type
func NewEventIP(data models.IP) {
	log.Debug("New event IP !")
	EventQueue <- Event{Type: EventIP, Timestamp: time.Now(), Data: data}
}

//NewEventDomain send new pre-filled event of the Domain type
func NewEventDomain(data models.Domain) {
	log.Debug("New event Domain !")
	EventQueue <- Event{Type: EventDomain, Timestamp: time.Now(), Data: data}
}

//NewEventPort send new pre-filled event of the Port type
func NewEventPort(data models.Port) {
	log.Debug("New event Port !")
	EventQueue <- Event{Type: EventPort, Timestamp: time.Now(), Data: data}
}

//NewEventURI send new pre-filled event of the URI type
func NewEventURI(data models.URI) {
	log.Debug("New event URI !")
	EventQueue <- Event{Type: EventURI, Timestamp: time.Now(), Data: data}
}
