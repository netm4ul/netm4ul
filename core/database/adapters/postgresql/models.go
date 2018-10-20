package postgresql

import (
	"net"
	"time"

	"github.com/go-pg/pg/orm"
	"github.com/jinzhu/gorm"
	"github.com/netm4ul/netm4ul/core/database/models"
)

/*
postgres model for Hop
*/
type pgHop struct {
	gorm.Model
	IP   net.IP
	Max  float32
	Min  float32
	Avg  float32
	IPId uint
}

func (pgHop) TableName() string {
	return "hops"
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
	gorm.Model
	Source      string
	Destination string
	ProjectID   uint
}

func (pgRoute) TableName() string {
	return "routes"
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
	gorm.Model
	Name string `gorm:"unique_index:idx_name_portid"`
	Code string

	PortID uint    `gorm:"unique_index:idx_name_portid"`
	Port   *pgPort // 1 to 1 relation
}

func (pgURI) TableName() string {
	return "uris"
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

type portToType struct {
	PortID     int
	PorttypeID int
}

func (portToType) TableName() string {
	return "port_to_types"
}

/*
	postgres model for Port
*/
type pgPort struct {
	gorm.Model
	// not using "inclusion" to simlify the models.Ports struct. But this fields MUST match.
	Number   int16  `gorm:"unique_index:idx_number_protocol_ipid"`
	Protocol string `gorm:"unique_index:idx_number_protocol_ipid"`
	Status   string
	Banner   string
	Type     string

	IP       pgIP
	IPId     uint         `gorm:"unique_index:idx_number_protocol_ipid"`
	PortType []pgPortType `gorm:"many2many:port_to_types;"` // a single port can have multiple can have multiple type.
}

func (pgPort) TableName() string {
	return "ports"
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
	gorm.Model
	Type        string
	Description string
}

func (pgPortType) TableName() string {
	return "port_types"
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
	gorm.Model
	Value     string `gorm:"unique_index:idx_value_network_project"`
	Network   string `sql:"default:'external'" gorm:"unique_index:idx_value_network_project"` // arbitrary value, default should be "external".
	ProjectID uint   `gorm:"unique_index:idx_value_network_project"`
	Project   pgProject
}

func (pgIP) TableName() string {
	return "ips"
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
	gorm.Model
	Name        string
	Description string
}

func (pgNetwork) TableName() string {
	return "networks"
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

type domainToIps struct {
	DomainID int
	IPId     int
}

func (domainToIps) TableName() string {
	return "domain_to_ips"
}

/*
	postgres model for Domain
*/
type pgDomain struct {
	gorm.Model
	Name      string `gorm:"unique_index:idx_name_project"`
	IP        []pgIP `gorm:"many2many:domain_to_ips;"`
	ProjectID uint   `gorm:"unique_index:idx_name_project"`
	Project   pgProject
}

func (pgDomain) TableName() string {
	return "domains"
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
	gorm.Model
	Name        string `gorm:"unique_index:idx_name"`
	Description string
	IPS         []*pgIP
}

func (pgProject) TableName() string {
	return "projects"
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

type userToProject struct {
	UserID    int
	ProjectID int
}

func (userToProject) TableName() string {
	return "user_to_projects"
}

/*
	postgres model for User
*/
type pgUser struct {
	gorm.Model
	Name     string `gorm:"unique_index:idx_name"`
	Password string
	Token    string
	Projects []*pgProject `gorm:"many2many:user_to_projects;"`
}

func (pgUser) TableName() string {
	return "users"
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
	gorm.Model
	Content    string
	ModuleName string

	ProjectID uint
	Project   *pgProject
}

func (pgRaw) TableName() string {
	return "raws"
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
