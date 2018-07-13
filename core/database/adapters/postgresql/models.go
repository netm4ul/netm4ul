package postgresql

import (
	"github.com/netm4ul/netm4ul/core/database/models"
)

type pgHop struct {
	models.Hop
	ID int
}

type pgRoute struct {
	models.Route
	ID int
}

type pgURI struct {
	models.URI
	ID int
}

type pgPort struct {
	models.Port
	ID int
}

type pgPortType struct {
	models.PortType
	ID int
}

type pgIP struct {
	models.IP
	ID int
}

type pgDomain struct {
	models.Domain
	ID int
}

type pgProject struct {
	models.Project
	ID int
}

type pgUser struct {
	models.User
	ID int
}
