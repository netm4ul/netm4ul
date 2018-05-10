package jsondb_test

import (
	"testing"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/adapters/jsondb"
)

var (
	BaseDir = "./test_files"
)

func TestJsonDB_GetRawModule(t *testing.T) {

	c := config.ConfigToml{
		Database: config.Database{
			DatabaseType: "JsonDB",
			IP:           "localhost",
			User:         "test",
			Password:     "test",
		},
	}

	j := jsondb.InitDatabase(&c)

	j.BaseDir = BaseDir
	j.RawPathFmt = j.BaseDir + "/raw-%s-%s.json"
	j.RawGlob = j.BaseDir + "/raw-"
	j.ResultPathFmt = j.BaseDir + "/project-%s.json"
	j.ProjectGlob = j.BaseDir + "/project-*"

	data, err := j.GetRawModule("netm4ul", "Traceroute")

	if err != nil {
		t.Error(err)
	}

	if _, ok := data["1525815067562763181"]; !ok {
		t.Error("data[\"1525815067562763181\"] does not exist !")
	}
	//TODO : Check result of GetRawModule
}
