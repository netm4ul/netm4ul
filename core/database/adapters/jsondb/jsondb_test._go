package jsondb_test

import (
	"strconv"
	"testing"

	"github.com/netm4ul/netm4ul/tests"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/adapters/jsondb"
	"github.com/netm4ul/netm4ul/core/database/models"
)

var (
	BaseDir = "./test_files"
	cfg     config.ConfigToml
	jdb     *jsondb.JsonDB
)

func init() {

	cfg = config.ConfigToml{
		Database: config.Database{
			DatabaseType: "JsonDB",
			IP:           "localhost",
			User:         "test",
			Password:     "test",
		},
	}

	jdb = jsondb.InitDatabase(&cfg)

	jdb.BaseDir = BaseDir
	jdb.RawPathFmt = jdb.BaseDir + "/raw-%s-%s.json"
	jdb.RawGlob = jdb.BaseDir + "/raw-"
	jdb.ResultPathFmt = jdb.BaseDir + "/project-%s.json"
	jdb.ProjectGlob = jdb.BaseDir + "/project-*"
}

func TestJsonDB_CreateOrUpdateProject(t *testing.T) {
	project := models.Project{Name: tests.NormalProject.Name, Description: tests.NormalProject.Description}

	err := jdb.CreateOrUpdateProject(project)
	if err != nil {
		t.Fatalf("Could not create or update project : %s", project.Name)
	}

	p, err := jdb.GetProject(tests.NormalProject.Name)
	if err != nil {
		t.Errorf("Could not get project %s : %s", tests.NormalProject.Name, err)
	}

	if p.Name != tests.NormalProject.Name {
		t.Errorf("Bad project name, expected %s, got %s", p.Name, tests.NormalProject.Name)
	}

	if p.Description != tests.NormalProject.Description {
		t.Errorf("Bad project description, expected %s, got %s", p.Description, tests.NormalProject.Description)
	}
}

func TestJsonDB_CreateOrUpdateIP(t *testing.T) {
	ip := models.IP{Value: tests.NormalProject.IPs[0].Value}
	err := jdb.CreateOrUpdateIP(tests.NormalProject.Name, ip)
	if err != nil {
		t.Errorf("Could not create or update IP : %s", ip.Value)
	}
	ips, err := jdb.GetIPs(tests.NormalProject.Name)
	if err != nil {
		t.Fatalf("Could not get IPs for project : %s", tests.NormalProject.Name)
	}
	if len(ips) == 0 {
		t.Fatalf("Didn't get any IP")
	}
	if ips[0].Value != tests.NormalProject.IPs[0].Value {
		t.Errorf("Read bad ip address, expected %s, got %s", tests.NormalProject.IPs[0].Value, ips[0].Value)
	}
}

func TestJsonDB_CreateOrUpdatePort(t *testing.T) {

	port := tests.NormalProject.IPs[0].Ports[0]
	port.URIs = nil
	err := jdb.CreateOrUpdatePort(tests.NormalProject.Name, tests.NormalProject.IPs[0].Value, port)
	if err != nil {
		t.Errorf("Could not create or update Port : %+v", port)
	}

	ports, err := jdb.GetPorts(tests.NormalProject.Name, tests.NormalProject.IPs[0].Value)
	if err != nil {
		t.Fatalf("Could not get ports for project : %s", tests.NormalProject.Name)
	}

	if len(ports) == 0 {
		t.Fatalf("Didn't get any port")
	}

	var gotPort models.Port
	found := false
	for _, p := range ports {
		if p.Number == port.Number && p.Protocol == port.Protocol {
			gotPort = p
			found = true
		}
	}

	if !found {
		t.Fatal("Could not match any port !")
	}

	if gotPort.Banner != port.Banner {
		t.Errorf("Bad banner for port, expected %s got %s", port.Banner, gotPort.Banner)
	}

	if gotPort.Number != port.Number {
		t.Errorf("Bad Number for port, expected %d got %d", port.Number, gotPort.Number)
	}

	if gotPort.Protocol != port.Protocol {
		t.Errorf("Bad Protocol for port, expected %s got %s", port.Protocol, gotPort.Protocol)
	}

	if gotPort.Status != port.Status {
		t.Errorf("Bad Status for port, expected %s got %s", port.Status, gotPort.Status)
	}

	if gotPort.Type != port.Type {
		t.Errorf("Bad Type for port, expected %s got %s", port.Type, gotPort.Type)
	}
}

func TestJsonDB_CreateOrUpdateURI(t *testing.T) {
	project := tests.NormalProject.Name
	ip := tests.NormalProject.IPs[0].Value
	port := strconv.Itoa(int(tests.NormalProject.IPs[0].Ports[0].Number))
	uri := tests.NormalProject.IPs[0].Ports[0].URIs[0]

	err := jdb.CreateOrUpdateURI(project, ip, port, uri)

	if err != nil {
		t.Errorf("Could not create or update URI : %s", err)
	}
	uris, err := jdb.GetURIs(project, ip, port)
	if err != nil {
		t.Fatalf("Could not get uris for project : %s", err)
	}

	if len(uris) == 0 {
		t.Fatalf("Didn't get any URIs")
	}

}

func TestJsonDB_GetRawModule(t *testing.T) {

	data, err := jdb.GetRawModule("netm4ul", "test")
	if err != nil {
		t.Error(err)
	}

	if _, ok := data["1525815067562763181"]; !ok {
		t.Error("data[\"1525815067562763181\"] does not exist !")
	}
}
