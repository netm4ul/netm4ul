package jsondb_test

import (
	"testing"

	"github.com/netm4ul/netm4ul/tests"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/adapters/jsondb"
	"github.com/netm4ul/netm4ul/core/database/models"
)

const (
	testProjectName = "Test project"
	testProjectDesc = "Test description"
	testIPValue     = "1.2.3.4"
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
	project := models.Project{Name: testProjectName, Description: testProjectDesc}

	err := jdb.CreateOrUpdateProject(project)
	if err != nil {
		t.Fatalf("Could not create or update project : %s", project.Name)
	}

	p, err := jdb.GetProject(testProjectName)
	if err != nil {
		t.Errorf("Could not get project %s : %s", testProjectName, err)
	}

	if p.Name != testProjectName {
		t.Errorf("Bad project name, expected %s, got %s", p.Name, testProjectName)
	}

	if p.Description != testProjectDesc {
		t.Errorf("Bad project description, expected %s, got %s", p.Description, testProjectDesc)
	}
}
func TestJsonDB_CreateOrUpdateIP(t *testing.T) {
	ip := models.IP{Value: testIPValue}
	err := jdb.CreateOrUpdateIP(testProjectName, ip)
	if err != nil {
		t.Errorf("Could not create or update IP : %s", ip.Value)
	}
	ips, err := jdb.GetIPs(testProjectName)
	if err != nil {
		t.Fatalf("Could not get IPs for project : %s", testProjectName)
	}
	if len(ips) == 0 {
		t.Fatalf("Didn't get any IP")
	}
	if ips[0].Value != testIPValue {
		t.Errorf("Read bad ip address, expected %s, got %s", testIPValue, ips[0].Value)
	}
}

func TestJsonDB_CreateOrUpdatePort(t *testing.T) {

	port := tests.NormalProject.IPs[0].Ports[0]
	port.URIs = nil
	err := jdb.CreateOrUpdatePort(testProjectName, testIPValue, port)
	if err != nil {
		t.Errorf("Could not create or update Port : %+v", port)
	}

	ports, err := jdb.GetPorts(testProjectName, testIPValue)
	if err != nil {
		t.Fatalf("Could not get ports for project : %s", testProjectName)
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

func TestJsonDB_GetRawModule(t *testing.T) {

	data, err := jdb.GetRawModule("netm4ul", "test")
	if err != nil {
		t.Error(err)
	}

	if _, ok := data["1525815067562763181"]; !ok {
		t.Error("data[\"1525815067562763181\"] does not exist !")
	}
}
