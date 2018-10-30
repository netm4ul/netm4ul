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
	//TODO
	// everyhting !
	var p models.Project
	var err error
	var data [][]string

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "Ports"})

	if projectName == "" {
		log.Fatalln("No project provided")
		// exit
	}

	p, err = requester.GetProject(projectName, s)
	if err != nil {
		log.Errorf("Can't get project %s : %s", projectName, err.Error())
	}

	log.Debugf("Project : %+v", p)

	table.AppendBulk(data)
	table.Render()
}

func PrintProjectsInfo(s *session.Session) {
	var err error
	var data [][]string

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Project", "Description", "# IPs", "Last update"})

	// get list of projects
	listOfProjects, err := requester.GetProjects(s)
	if err != nil {
		log.Errorf("Can't get projects list : %s", err.Error())
	}

	// build array of array for the table !
	for _, p := range listOfProjects {
		if s.Verbose {
			log.Infof("p : %+v", p)
		}
		data = append(data, []string{p.Name, p.Description, strconv.Itoa(int(p.UpdatedAt.Unix()))})
	}

	table.AppendBulk(data)
	table.Render()
}
