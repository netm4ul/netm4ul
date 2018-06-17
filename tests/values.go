package tests

import (
	"github.com/netm4ul/netm4ul/core/database/models"
)

var (
	NormalProject  models.Project
	EmptyProject   models.Project
	NormalRaw      models.Raws
	NormalProjects []models.Project
	NormalRaws     map[string]models.Raws
	NormalUser     models.User
)

func init() {
	NormalProject = models.Project{
		ID:          "0",
		Name:        "Test project",
		Description: "Test description",
		IPs: []models.IP{
			models.IP{
				ID:    "1",
				Value: "1.1.1.1", Ports: []models.Port{
					models.Port{
						ID:       "2",
						Banner:   "Test banner",
						Number:   80,
						Protocol: "tcp",
						Status:   "open",
						Type:     "web",
						URIs: []models.URI{
							models.URI{
								ID:   "3",
								Code: "200",
								Name: "Test URI",
							},
							models.URI{
								ID:   "4",
								Code: "404",
								Name: "Test not found URI",
							},
							models.URI{
								ID:   "5",
								Code: "500",
								Name: "Test server error URI",
							},
						},
					},
					models.Port{
						ID:       "6",
						Banner:   "Test banner 2",
						Number:   22,
						Protocol: "tcp",
						Status:   "open",
						Type:     "ssh",
						URIs:     []models.URI{}, // empty uri for ssh
					},
				},
			},
			models.IP{
				ID:    "7",
				Value: "2.2.2.2",
				Ports: []models.Port{}, // empty ports
			},
		},
		Domains: []models.Domain{
			models.Domain{
				Name: "domain.tld",
				SubDomains: []models.Domain{
					models.Domain{Name: "subdomain"},
					models.Domain{Name: "subdomain2"},
				},
			},
			models.Domain{
				Name: "another.tld",
				SubDomains: []models.Domain{
					models.Domain{Name: "anothersub"},
					models.Domain{Name: "anothersub2"},
				},
			},
		},
	}

	NormalRaw = models.Raws{
		NormalProject.Name: map[string]interface{}{
			"Test module array":   []string{"test value", "test value 2"},
			"Test module string":  "test value",
			"Test module integer": 18,
		},
	}

	NormalProjects = []models.Project{
		NormalProject,
		EmptyProject,
	}

	NormalRaws = map[string]models.Raws{
		NormalProject.Name: NormalRaw,
	}

	NormalUser = models.User{
		ID:        "1",
		Name:      "TestUser",
		Password:  "$2y$10$Fu4hg./ZybmFjiPxIpEOROGwQhF3sfwakddzlWFtV.I3rJu6sfy/2", // Test password
		Token:     "testtoken",
		CreatedAt: 123,
		UpdatedAt: 456,
	}
}
