package postgresql

import (
	"github.com/netm4ul/netm4ul/core/database/models"
)

type pgHop struct {
	tableName struct{} `sql:"alias:hops"`
	models.Hop
	ID int
}

func (p *pgHop) ToModel() models.Hop {
	hop := models.Hop{
		Avg: p.Avg,
		Min: p.Min,
		Max: p.Max,
		IP:  p.IP,
	}
	return hop
}

type pgRoute struct {
	tableName struct{} `sql:"alias:routes"`
	models.Route
	ID int
}

func (p *pgRoute) ToModel() models.Route {
	route := models.Route{
		Source:      p.Source,
		Destination: p.Destination,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return route
}

type pgURI struct {
	tableName struct{} `sql:"alias:uris"`
	models.URI
	ID int
}

func (p *pgURI) ToModel() models.URI {
	uri := models.URI{
		Name: p.Name,
		Code: p.Code,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return uri
}

type pgPort struct {
	tableName struct{} `sql:"alias:ports"`
	models.Port
	ID int
}

func (p *pgPort) ToModel() models.Port {
	port := models.Port{
		Type:     p.Type,
		Status:   p.Status,
		Protocol: p.Protocol,
		Number:   p.Number,
		Banner:   p.Banner,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return port
}

type pgPortType struct {
	tableName struct{} `sql:"alias:porttypes"`
	models.PortType
	ID int
}

func (p *pgPortType) ToModel() models.PortType {
	porttype := models.PortType{
		Type:        p.Type,
		Description: p.Description,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return porttype
}

type pgIP struct {
	tableName struct{} `sql:"alias:ips"`
	models.IP
	ID int
}

func (p *pgIP) ToModel() models.IP {
	ip := models.IP{
		Value:   p.Value,
		Network: p.Network,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return ip
}

type pgDomain struct {
	tableName struct{} `sql:"alias:domains"`
	models.Domain
	ID int
}

func (p *pgDomain) ToModel() models.Domain {
	domain := models.Domain{
		Name: p.Name,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return domain
}

type pgProject struct {
	tableName struct{} `sql:"alias:projects"`
	models.Project
	ID  int
	IPS []pgIP `pg:"many2many:project_to_ips"`
}

func (p *pgProject) ToModel() models.Project {
	project := models.Project{
		Name:        p.Name,
		Description: p.Description,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return project
}

type pgUser struct {
	tableName struct{} `sql:"alias:users"`
	models.User
	ID       int
	Projects []pgProject `pg:"many2many:users_to_projects"`
}

func (p *pgUser) ToModel() models.User {
	user := models.User{
		Name:     p.Name,
		Password: p.Password,
		Token:    p.Token,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return user
}
