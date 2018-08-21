package postgresql

import (
	"time"

	"github.com/go-pg/pg/orm"
	"github.com/netm4ul/netm4ul/core/database/models"
)

/*
postgres model for Hop
*/
type pgHop struct {
	tableName struct{} `sql:"hops"`
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

func (p *pgHop) FromModel(h models.Hop) {
	p.Avg = h.Avg
	p.Min = h.Min
	p.Max = h.Max
	p.IP = h.IP
}

/*
postgres model for Route
*/
type pgRoute struct {
	tableName struct{} `sql:"routes"`
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

func (p *pgRoute) FromModel(r models.Route) {
	p.Source = r.Source
	p.Destination = r.Destination

	p.CreatedAt = r.CreatedAt
	p.UpdatedAt = r.UpdatedAt
}

func (p *pgRoute) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for URI
*/
type pgURI struct {
	tableName struct{} `sql:"uris"`
	models.URI
	ID   int
	Port *pgPort // 1 to 1 relation
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

func (p *pgURI) FromModel(uri models.URI) {
	p.Name = uri.Name
	p.Code = uri.Code

	p.CreatedAt = uri.CreatedAt
	p.UpdatedAt = uri.UpdatedAt
}

func (p *pgURI) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for Port
*/
type pgPort struct {
	tableName struct{} `sql:"ports"`
	models.Port
	ID       int
	IP       *pgIP
	PortType *pgPortType `pg:",many2many:port_to_types"`
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

func (p *pgPort) FromModel(pt models.Port) {
	p.Banner = pt.Banner
	p.Number = pt.Number
	p.Protocol = pt.Protocol
	p.Status = pt.Status
	p.Type = pt.Type

	p.CreatedAt = pt.CreatedAt
	p.UpdatedAt = pt.UpdatedAt
}

func (p *pgPort) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for Port type
*/
type pgPortType struct {
	tableName struct{} `sql:"porttypes"`
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

func (p *pgPortType) FromModel(pt models.PortType) {
	p.Type = pt.Type
	p.Description = pt.Description

	p.CreatedAt = pt.CreatedAt
	p.UpdatedAt = pt.UpdatedAt
}

func (p *pgPortType) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for IP
*/
type pgIP struct {
	tableName struct{} `sql:"ips"`
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

func (p *pgIP) FromModel(ip models.IP) {
	p.Value = ip.Value
	p.Network = ip.Network

	p.CreatedAt = ip.CreatedAt
	p.UpdatedAt = ip.UpdatedAt
}

func (p *pgIP) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for Network
*/
type pgNetwork struct {
	tableName struct{} `sql:"Networks"`
	models.Network
	ID int
}

func (p *pgNetwork) ToModel() models.Network {
	Network := models.Network{
		Name:        p.Name,
		Description: p.Description,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return Network
}

func (p *pgNetwork) FromModel(Network models.Network) {
	p.Name = Network.Name
	p.Description = Network.Description

	p.CreatedAt = Network.CreatedAt
	p.UpdatedAt = Network.UpdatedAt
}

func (p *pgNetwork) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for Domain
*/
type pgDomain struct {
	tableName struct{} `sql:"domains"`
	models.Domain
	ID int
	IP []*pgIP `pg:",many2many:domain_to_ips"`
}

func (p *pgDomain) ToModel() models.Domain {
	domain := models.Domain{
		Name: p.Name,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	return domain
}

func (p *pgDomain) FromModel(d models.Domain) {
	p.Name = d.Name

	p.CreatedAt = d.CreatedAt
	p.UpdatedAt = d.UpdatedAt
}

func (p *pgDomain) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for Project
*/
type pgProject struct {
	tableName struct{} `sql:"projects"`
	models.Project
	ID  int
	IPS []*pgIP
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

func (p *pgProject) FromModel(proj models.Project) {
	p.Name = proj.Name
	p.Description = proj.Description

	p.CreatedAt = proj.CreatedAt
	p.UpdatedAt = proj.UpdatedAt
}

func (p *pgProject) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for User
*/
type pgUser struct {
	tableName struct{} `sql:"users"`
	models.User
	ID       int
	Projects []*pgProject `pg:"many2many:users_to_projects"`
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

func (p *pgUser) FromModel(u models.User) {
	p.Name = u.Name
	p.Password = u.Password
	p.Token = u.Token

	p.CreatedAt = u.CreatedAt
	p.UpdatedAt = u.UpdatedAt
}

func (p *pgUser) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}

/*
	postgres model for Raw data
*/
type pgRaw struct {
	tableName struct{} `sql:"raws"`
	models.Raw
	ID      int
	Project *pgProject
}

func (p *pgRaw) ToModel() models.Raw {
	raws := models.Raw{
		Content:    p.Content,
		ModuleName: p.ModuleName,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	return raws
}

func (p *pgRaw) FromModel(r models.Raw) {
	p.Content = r.Content
	p.ModuleName = r.ModuleName

	p.CreatedAt = r.CreatedAt
	p.UpdatedAt = r.UpdatedAt
}

func (p *pgRaw) BeforeInsert(db orm.DB) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now()
	}
	return nil
}
