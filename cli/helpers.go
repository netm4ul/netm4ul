package cmd

import (
	"github.com/netm4ul/netm4ul/core/database/models"
	"io/ioutil"
	"net"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/session"

	"github.com/pkg/errors"
)

/*
parseModules gets all the modules configured for this session.
Return :
	- slice of all the modules name (string)
	- error is anything unexpected happens
*/
func parseModules(modules []string, s *session.Session) ([]string, error) {

	if len(modules) == 0 {
		return nil, errors.New("Could not parse modules")
	}

	for _, name := range modules {
		_, ok := s.Config.Modules[name]
		if !ok {
			return nil, errors.New("Could not find module : " + name)
		}
	}

	return modules, nil
}

func addModules(mods []string, s *session.Session) {

	found := false
	for _, mod := range mods {
		for cmodname, cmod := range s.Config.Modules {
			if cmodname == mod && cmod.Enabled {
				found = true
			}
		}
		if !found {
			s.Config.Modules[mod] = config.Module{Enabled: true}
		}
	}
}

func parseTargets(targets []string) ([]communication.Input, error) {

	var inputs []communication.Input
	var input communication.Input

	if len(targets) == 0 {
		return []communication.Input{}, errors.New("Not target found")
	}

	// loop on each targets
	for _, target := range targets {

		ip, ipNet, err := net.ParseCIDR(target)

		// if this is a domain
		if err != nil {
			ips, err := net.LookupIP(target)

			if err != nil {
				return []communication.Input{}, errors.New("Could lookup address : " + target + ", " + err.Error())
			}

			if ips == nil {
				return []communication.Input{}, errors.New("Could not resolve :" + target)
			}

			// convert ips to strings
			for _, ip := range ips {
				input = communication.Input{Domain: models.Domain{Name: target}, IP: models.IP{Value: ip.String()}}
				inputs = append(inputs, input)
			}

		} else {
			// if this is an ip
			// check if ip is specified (not :: or 0.0.0.0)
			if ip.IsUnspecified() {
				return []communication.Input{}, errors.New("Target ip is Unspecified (0.0.0.0 or ::)")
			}

			// check if ip isn't loopback
			if ip.IsLoopback() {
				return []communication.Input{}, errors.New("Target ip is loopback address")
			}

			// IP Range (CIDR)
			if ipNet != nil {
				h, err := hosts(target)
				if err != nil {
					return []communication.Input{}, errors.New("Target ip range is invalid (" + err.Error() + ")")
				}
				for _, host := range h {
					input = communication.Input{IP: models.IP{Value: host.String()}}
					inputs = append(inputs, input)
				}
			}
		}
	}

	return inputs, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func hosts(cidr string) ([]net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip)
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func getGlobalModulesList() ([]string, error) {
	res := []string{}

	exploitsModules, err := getModulesList("exploit")
	if err != nil {
		return nil, errors.New("Could not load exploit modules : " + err.Error())
	}

	reconsModules, err := getModulesList("recon")
	if err != nil {
		return nil, errors.New("Could not load recon modules : " + err.Error())
	}

	reportsModules, err := getModulesList("report")
	if err != nil {
		return nil, errors.New("Could not load report modules : " + err.Error())
	}

	res = append(res, exploitsModules...)
	res = append(res, reconsModules...)
	res = append(res, reportsModules...)
	return res, nil
}

func getModulesList(modType string) ([]string, error) {
	files, err := ioutil.ReadDir("./modules/" + modType)
	if err != nil {
		return nil, err
	}

	res := []string{}
	for _, f := range files {
		if f.IsDir() {
			res = append(res, f.Name())
		}
	}
	return res, nil
}

//TOFIX : must be a better way
func setDefaultValues(cfg *config.ConfigToml) {

	//Algorithm
	if cfg.Algorithm.Name == "" {
		cfg.Algorithm.Name = defaultAlgorithm
	}

	//API
	if cfg.API.Port == 0 {
		cfg.API.Port = defaultAPIPort
	}
	if cfg.API.User == "" {
		cfg.API.User = defaultAPIUser
	}

	//DATABASE
	if cfg.Database.Database == "" {
		cfg.Database.Database = defaultDBname
	}
	if cfg.Database.DatabaseType == "" {
		cfg.Database.DatabaseType = defaultDBType
	}
	if cfg.Database.IP == "" {
		cfg.Database.IP = defaultDBIP
	}
	if cfg.Database.User == "" {
		cfg.Database.User = defaultDBSetupUser
	}
	if cfg.Database.Password == "" {
		cfg.Database.Password = defaultDBSetupPassword
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = defaultDBPort
	}

	//Project
	if cfg.Project.Name == "" {
		cfg.Project.Name = defaultProjectName
	}
	if cfg.Project.Description == "" {
		cfg.Project.Description = defaultProjectDescription
	}

	//Modules
	if cfg.Modules == nil {
		cfg.Modules = make(map[string]config.Module, 0)
	}
}
