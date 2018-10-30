package ui

import (
	"fmt"
	"os"
	"strconv"

	"github.com/netm4ul/netm4ul/cli/requester"
	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/client"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

//PrintVersion Prints the version of all the components : The server, the Client, and the HTTP API
func PrintVersion(s *session.Session) {
	fmt.Printf("Version :\n - Server : %s\n - Client : %s\n - HTTP API : %s\n", server.Version, client.Version, api.Version)
}

func PrintProjectInfo(projectName string, s *session.Session) {
	var p models.Project
	var err error
	var data [][]string

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "Created at", "Updated at"})

	if projectName == "" {
		log.Fatalln("No project provided")
	}

	p, err = requester.GetProject(projectName, s)
	if err != nil {
		log.Errorf("Can't get project %s : %s", projectName, err.Error())
	}
	data = append(data, []string{p.Name, p.Description, p.CreatedAt.String(), p.UpdatedAt.String()})
	table.AppendBulk(data)
	table.Render()
}

func PrintProjectsInfo(s *session.Session) {
	var err error
	var data [][]string

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Project", "Description", "# IPs", "# Domains", "# Ports", "# URI", "Created at", "Last update"})

	// get list of projects
	listOfProjects, err := requester.GetProjects(s)
	if err != nil {
		log.Errorf("Can't get projects list : %s", err.Error())
	}

	// build array of array for the table !
	for _, p := range listOfProjects {
		portList := []models.Port{}
		uriList := []models.URI{}

		ipList, err := requester.GetIPs(p.Name, s)
		if err != nil {
			log.Fatalf("Couldn't get ips list : %s", err)
		}

		domainList, err := requester.GetDomains(p.Name, s)
		if err != nil {
			log.Fatalf("Couldn't get ips list : %s", err)
		}
		for _, ip := range ipList {
			pl, err := requester.GetPorts(p.Name, ip.Value, s)
			if err != nil {
				log.Fatalf("Couldn't get ips list : %s", err)
			}
			portList = append(portList, pl...)

			for _, port := range pl {
				ul, err := requester.GetURIs(p.Name, ip.Value, strconv.Itoa(int(port.Number)), s)
				if err != nil {
					log.Fatalf("Couldn't get ips list : %s", err)
				}
				uriList = append(uriList, ul...)
			}
		}

		data = append(data, []string{
			p.Name,
			p.Description,
			strconv.Itoa(len(ipList)),
			strconv.Itoa(len(domainList)),
			strconv.Itoa(len(portList)),
			strconv.Itoa(len(uriList)),
			p.CreatedAt.String(),
			p.UpdatedAt.String(),
		})
	}

	table.AppendBulk(data)
	table.Render()
}
