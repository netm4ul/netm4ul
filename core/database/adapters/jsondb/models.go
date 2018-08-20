package jsondb

import (
	"github.com/netm4ul/netm4ul/core/database/models"
)

/*
JSON model for Hop
*/
type jsonHop struct {
	models.Hop
	ID int
}

func (jm *jsonHop) ToModel() models.Hop {
	hop := models.Hop{
		Avg: jm.Avg,
		Min: jm.Min,
		Max: jm.Max,
		IP:  jm.IP,
	}
	return hop
}

func (jm *jsonHop) FromModel(h models.Hop) {
	jm.Avg = h.Avg
	jm.Min = h.Min
	jm.Max = h.Max
	jm.IP = h.IP
}

/*
JSON model for Route
*/
type jsonRoute struct {
	models.Route
	ID int
}

func (jm *jsonRoute) ToModel() models.Route {
	route := models.Route{
		Source:      jm.Source,
		Destination: jm.Destination,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return route
}

func (jm *jsonRoute) FromModel(r models.Route) {
	jm.Source = r.Source
	jm.Destination = r.Destination

	jm.CreatedAt = r.CreatedAt
	jm.UpdatedAt = r.UpdatedAt
}

/*
	JSON model for URI
*/
type jsonURI struct {
	models.URI
	ID   int
	Port *jsonPort // 1 to 1 relation
}

func (jm *jsonURI) ToModel() models.URI {
	uri := models.URI{
		Name: jm.Name,
		Code: jm.Code,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return uri
}

func (jm *jsonURI) FromModel(uri models.URI) {
	jm.Name = uri.Name
	jm.Code = uri.Code

	jm.CreatedAt = uri.CreatedAt
	jm.UpdatedAt = uri.UpdatedAt
}

/*
	JSON model for Port
*/
type jsonPort struct {
	models.Port
	ID       int
	PortType jsonPortType
	URIs     []jsonURI
}

func (jm *jsonPort) ToModel() models.Port {
	port := models.Port{
		Type:     jm.Type,
		Status:   jm.Status,
		Protocol: jm.Protocol,
		Number:   jm.Number,
		Banner:   jm.Banner,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return port
}

func (jm *jsonPort) FromModel(pt models.Port) {
	jm.Banner = pt.Banner
	jm.Number = pt.Number
	jm.Protocol = pt.Protocol
	jm.Status = pt.Status
	jm.Type = pt.Type

	jm.CreatedAt = pt.CreatedAt
	jm.UpdatedAt = pt.UpdatedAt
}

/*
	JSON model for Port type
*/
type jsonPortType struct {
	models.PortType
	ID int
}

func (jm *jsonPortType) ToModel() models.PortType {
	porttype := models.PortType{
		Type:        jm.Type,
		Description: jm.Description,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return porttype
}

func (jm *jsonPortType) FromModel(pt models.PortType) {
	jm.Type = pt.Type
	jm.Description = pt.Description

	jm.CreatedAt = pt.CreatedAt
	jm.UpdatedAt = pt.UpdatedAt
}

/*
	JSON model for IP
*/
type jsonIP struct {
	models.IP
	ID    int
	Ports []jsonPort
}

func (jm *jsonIP) ToModel() models.IP {
	ip := models.IP{
		Value:   jm.Value,
		Network: jm.Network,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return ip
}

func (jm *jsonIP) FromModel(ip models.IP) {
	jm.Value = ijm.Value
	jm.Network = ijm.Network

	jm.CreatedAt = ijm.CreatedAt
	jm.UpdatedAt = ijm.UpdatedAt
}

/*
	JSON model for Network
*/
type jsonNetwork struct {
	models.Network
	ID int
}

func (jm *jsonNetwork) ToModel() models.Network {
	Network := models.Network{
		Name:        jm.Name,
		Description: jm.Description,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return Network
}

func (jm *jsonNetwork) FromModel(Network models.Network) {
	jm.Name = Network.Name
	jm.Description = Network.Description

	jm.CreatedAt = Network.CreatedAt
	jm.UpdatedAt = Network.UpdatedAt
}

/*
	JSON model for Domain
*/
type jsonDomain struct {
	models.Domain
	ID int
	IP []*jsonIP
}

func (jm *jsonDomain) ToModel() models.Domain {
	domain := models.Domain{
		Name: jm.Name,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return domain
}

func (jm *jsonDomain) FromModel(d models.Domain) {
	jm.Name = d.Name

	jm.CreatedAt = d.CreatedAt
	jm.UpdatedAt = d.UpdatedAt
}

/*
	JSON model for Project
*/
type jsonProject struct {
	models.Project
	ID  int
	IPS []jsonIP
}

func (jm *jsonProject) ToModel() models.Project {
	project := models.Project{
		Name:        jm.Name,
		Description: jm.Description,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return project
}

func (jm *jsonProject) FromModel(proj models.Project) {
	jm.Name = proj.Name
	jm.Description = proj.Description

	jm.CreatedAt = proj.CreatedAt
	jm.UpdatedAt = proj.UpdatedAt
}

/*
	JSON model for User
*/
type jsonUser struct {
	models.User
	ID       int
	Projects []jsonProject
}

func (jm *jsonUser) ToModel() models.User {
	user := models.User{
		Name:     jm.Name,
		Password: jm.Password,
		Token:    jm.Token,

		CreatedAt: jm.CreatedAt,
		UpdatedAt: jm.UpdatedAt,
	}
	return user
}

func (jm *jsonUser) FromModel(u models.User) {
	jm.Name = u.Name
	jm.Password = u.Password
	jm.Token = u.Token

	jm.CreatedAt = u.CreatedAt
	jm.UpdatedAt = u.UpdatedAt
}

/*
	JSON model for Raw data
*/
type jsonRaws struct {
	models.Raws
	ID      int
	Project *jsonProject
}

func (jm *jsonRaws) ToModel() models.Raws {
	raws := models.Raws{
		Content:    jm.Content,
		ModuleName: jm.ModuleName,
		CreatedAt:  jm.CreatedAt,
		UpdatedAt:  jm.UpdatedAt,
	}
	return raws
}

func (jm *jsonRaws) FromModel(r models.Raws) {
	jm.Content = r.Content
	jm.ModuleName = r.ModuleName

	jm.CreatedAt = r.CreatedAt
	jm.UpdatedAt = r.UpdatedAt
}
