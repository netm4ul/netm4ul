package report

import (
	"strings"

	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/report/text"
)

var (
	//Reporter is the list of report's available ["pdf","text"...]
	Reporter map[string]modules.Report
)

//LoadReports is the init function
func LoadReports() {
	Reporter = make(map[string]modules.Report, 0)
	t := text.NewReport()
	Register(t)
}

//Register a new report
func Register(r modules.Report) {
	reportName := strings.ToLower(r.Name())
	Reporter[reportName] = r
}
