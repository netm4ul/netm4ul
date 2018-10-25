package tests

import (
	"time"

	"github.com/netm4ul/netm4ul/core/database/models"
)

var (
	NormalProject  models.Project
	EmptyProject   models.Project
	NormalRaw      models.Raw
	NormalProjects []models.Project
	NormalIPs      []models.IP
	NormalPorts    []models.Port
	NormalURIs     []models.URI
	NormalDomains  []models.Domain
	NormalRaws     map[string][]models.Raw
	NormalUser     models.User
)

func init() {
	NormalProject = models.Project{
		Name:        "Test project",
		Description: "Test description",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	NormalIPs = []models.IP{
		{
			Value: "1.1.1.1",
			Network: "external",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Value: "2.2.2.2",
			Network: "internal",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	NormalPorts = []models.Port{
		{
			Banner:   "Test banner",
			Number:   80,
			Protocol: "tcp",
			Status:   "open",
			Type:     "web",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Banner:   "Test banner 2",
			Number:   22,
			Protocol: "tcp",
			Status:   "open",
			Type:     "ssh",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	NormalURIs = []models.URI{
		{
			Code: "200",
			Name: "noslash",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Code: "200",
			Name: "middle/slashlol",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Code: "200",
			Name: "/test/uri",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Code: "404",
			Name: "/test/not_found",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Code: "500",
			Name: "/test/server/error",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Code: "1337",
			Name: "Testing non HTTP request URI",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	NormalDomains = []models.Domain{
		{
			Name: "domain.tld",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name: "another.tld",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name: "sub1.another.tld",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	NormalRaw = models.Raw{
		ModuleName: "TestModule",
		Content:    "Test content woa {][@&~#{[[|[`]@^ùm!::!,:;,n=))1234567\nabcder 1231<x§:;>é&",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	NormalProjects = []models.Project{
		NormalProject,
		EmptyProject,
	}

	NormalRaws = map[string][]models.Raw{
		"Test": {
			NormalRaw,
			NormalRaw,
		},
	}

	NormalUser = models.User{
		Name:      "TestUser",
		Password:  "$2y$10$Fu4hg./ZybmFjiPxIpEOROGwQhF3sfwakddzlWFtV.I3rJu6sfy/2", // Test password
		Token:     "testtoken",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
