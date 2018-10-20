package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/netm4ul/netm4ul/core/security"

	"github.com/netm4ul/netm4ul/core/config"

	"github.com/netm4ul/netm4ul/core/database/models"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type PromptRes struct {
	Message      string
	DefaultValue string
}

const (
	passwordLen               = 20
	defaultConfigPath         = "netm4ul.conf"
	defaultDBSetupUser        = "postgres"
	defaultDBname             = "netm4ul"
	defaultDBIP               = "localhost"
	defaultDBPort             = uint16(5432)
	defaultDBType             = "postgresql"
	defaultAPIUser            = "user"
	defaultAPIIP              = "localhost"
	defaultAPIPort            = uint16(8080)
	defaultServerIP           = "localhost"
	defaultServerPort         = uint16(444)
	defaultServerUser         = "user"
	defaultTLS                = "y" // Yes/y No/n (case insensitive)
	defaultCreateUser         = "y" // Yes/y No/n (case insensitive)
	defaultModuleStatus       = "y" // Yes/y No/n (case insensitive)
	defaultAlgorithm          = "random"
	defaultProjectName        = "first"
	defaultProjectDescription = "Your first project"
	defaultPasswordWordlist   = "0"
	defaultDomainWordlist     = "0"
)

var (
	defaultDBSetupPassword, _ = GeneratePassword(passwordLen)
	defaultAPIPassword, _     = GeneratePassword(passwordLen)
	defaultServerPassword, _  = GeneratePassword(passwordLen)
	cliDBSetupUser            string
	cliDBSetupPassword        string
	cliDBSetupIP              string
	cliDBSetupPort            uint16
	cliDBSetupType            string
	cliServerPassword         string
	cliServerIP               string
	cliServerPort             uint16
	cliApiUser                string
	cliApiPassword            string
	cliApiIP                  string
	cliApiPort                uint16
	cliDisableTLS             bool
	cliAlgorithm              string
	cliWordlist               string
	skipDBSetup               bool
	skipServerSetup           bool
	skipApiSetup              bool
	skipModulesSetup          bool
	skipAlgorithmSetup        bool
	skipProjectSetup          bool
	skipWordlists             bool
)

var setupDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "NetM4ul setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := setupDB()
		if err != nil {
			log.Errorln("Could not setup the database : " + err.Error())
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

var setupAPICmd = &cobra.Command{
	Use:   "api",
	Short: "NetM4ul setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := setupAPI()
		if err != nil {
			log.Errorln("Could not setup the API : " + err.Error())
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

var setupServerCmd = &cobra.Command{
	Use:   "server",
	Short: "NetM4ul setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := setupServer()
		if err != nil {
			log.Errorln("Could not setup the Server : " + err.Error())
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

var setupAlgorithmCmd = &cobra.Command{
	Use:   "algorithm",
	Short: "NetM4ul setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := setupAlgorithm()
		if err != nil {
			log.Errorln("Could not setup the Algorithm : " + err.Error())
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

var setupWordlistsCmd = &cobra.Command{
	Use:   "wordlists",
	Short: "NetM4ul setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := setupWordlists()
		if err != nil {
			log.Errorln("Could not setup the Wordlists : " + err.Error())
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

var setupTLSCmd = &cobra.Command{
	Use:   "tls",
	Short: "NetM4ul setup",
	Run: func(cmd *cobra.Command, args []string) {
		err := setupAlgorithm()
		if err != nil {
			log.Errorln("Could not setup the Algorithm : " + err.Error())
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "NetM4ul setup",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()

		if CLISession.ConfigPath == "" {
			CLISession.ConfigPath = defaultConfigPath
		}
	},

	Run: func(cmd *cobra.Command, args []string) {

		var err error

		// Create a new config file if it doesn't exist.
		// Use the saveConfigFile so we don't need to copy an "example file" to copy and maintain.
		if _, err := os.Stat(CLISession.ConfigPath); err != nil {
			err := saveConfigFile()
			if err != nil {
				log.Fatalf("Could not create: %s", err.Error())
			}
		}

		CLISession.Config, err = loadExistingConfig()
		if err != nil {
			log.Fatalf("Could load the existing config file %s", err.Error())
		}

		if !skipProjectSetup {
			err = setupProject()
			if err != nil {
				log.Fatalf("Could not setup the project : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping Project setup")
		}

		if !skipDBSetup {
			err = setupDB()
			if err != nil {
				log.Fatalf("Could not setup the database : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping Database setup")
		}

		if !skipApiSetup {
			err = setupAPI()
			if err != nil {
				log.Fatalf("Could not setup the API : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping API setup")
		}

		if !skipServerSetup {
			err = setupServer()
			if err != nil {
				log.Fatalf("Could not setup the Server : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping Server setup")
		}

		if !skipModulesSetup {
			err = setupModules()
			if err != nil {
				log.Fatalf("Could not setup modules : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping modules setup")
		}

		if !skipAlgorithmSetup {
			err = setupAlgorithm()
			if err != nil {
				log.Fatalf("Could not setup the algorithm : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping algorithm setup")
		}

		if !skipWordlists {
			err = setupWordlists()
			if err != nil {
				log.Fatalf("Could not setup the wordlists : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping wordlists setup")
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

// prompt user for configuration parameters
func prompt(param string) (answer string) {
	// var text string
	var input string

	// Database parameters
	promptString := map[string]PromptRes{
		"dbdel":              {Message: color.RedString("[DANGER]") + "Do you want to delete the database (default : %s) [y/N]: ", DefaultValue: "N"},
		"dbuser":             {Message: "Database username (default : %s) : ", DefaultValue: CLISession.Config.Database.User},
		"dbpassword":         {Message: "Database password (generated : (" + color.RedString("%s") + ") : ", DefaultValue: CLISession.Config.Database.Password},
		"dbip":               {Message: "Database IP (default : %s) : ", DefaultValue: CLISession.Config.Database.IP},
		"dbport":             {Message: "Database Port (default : %s) : ", DefaultValue: strconv.Itoa(int(CLISession.Config.Database.Port))},
		"dbtype":             {Message: "Database type [postgres, jsondb, mongodb] (default : %s): ", DefaultValue: CLISession.Config.Database.DatabaseType},
		"dbname":             {Message: "Database name (default : %s): ", DefaultValue: CLISession.Config.Database.Database},
		"apiuser":            {Message: "API username (default : %s) : ", DefaultValue: CLISession.Config.API.User},
		"apipassword":        {Message: "API password (generated : (" + color.RedString("%s") + ") : ", DefaultValue: defaultAPIPassword},
		"apiport":            {Message: "API port (default : %s) : ", DefaultValue: strconv.Itoa(int(CLISession.Config.API.Port))},
		"serverip":           {Message: "Server IP (default : %s) : ", DefaultValue: CLISession.Config.Server.IP},
		"serverport":         {Message: "Server port (default : %s) : ", DefaultValue: strconv.Itoa(int(CLISession.Config.Server.Port))},
		"serverpassword":     {Message: "Server password (generated : (" + color.RedString("%s") + ") : ", DefaultValue: CLISession.Config.Server.Password},
		"usetls":             {Message: "Use TLS (default : %s) [Y/n]: ", DefaultValue: defaultTLS},
		"createuser":         {Message: "Create a new user (default : %s) [Y/n]: ", DefaultValue: defaultCreateUser},
		"algorithm":          {Message: "Load balancing algorithm (default : %s) : ", DefaultValue: CLISession.Config.Algorithm.Name},
		"projectname":        {Message: "Project name (default : %s) : ", DefaultValue: CLISession.Config.Project.Name},
		"projectdescription": {Message: "Project description (default : %s) : ", DefaultValue: CLISession.Config.Project.Description},
		"wordlist-password":  {Message: "Choose a password wordlist to download [default : %s]: \n", DefaultValue: defaultPasswordWordlist},
		"wordlist-domain":    {Message: "Choose a domain wordlist to download [default : %s]: \n", DefaultValue: defaultDomainWordlist},
	}

	fmt.Printf(promptString[param].Message, promptString[param].DefaultValue)
	fmt.Scanln(&input)
	if input == "" {
		return promptString[param].DefaultValue
	}

	return input
}

func setupProject() error {
	CLISession.Config.Project.Name = prompt("projectname")
	CLISession.Config.Project.Description = prompt("projectdescription")
	return nil
}

func setupModules() error {
	modules, err := getGlobalModulesList()
	if err != nil {
		return errors.New("Could not load modules list : " + err.Error())
	}

	var input string
	var isEnabled bool

	for _, m := range modules {
		fmt.Printf("Enable module '%s' [Y/n] : ", m)
		fmt.Scanln(&input)
		if input == "" {
			isEnabled, err = yesNo(defaultModuleStatus)
		} else {
			isEnabled, err = yesNo(input)
		}
		CLISession.Config.Modules[m] = config.Module{Enabled: isEnabled}
	}

	return nil
}

// setupDB => create user in db for future requests
func setupDB() error {

	var err error

	CLISession.Config.Database.IP = prompt("dbip")

	p, err := strconv.Atoi(prompt("dbport"))
	if err != nil {
		return err
	}
	CLISession.Config.Database.Port = uint16(p)
	CLISession.Config.Database.DatabaseType = prompt("dbtype")
	CLISession.Config.Database.User = prompt("dbuser")
	CLISession.Config.Database.Password = prompt("dbpassword")
	CLISession.Config.Database.Database = prompt("dbname")

	db, err := database.NewDatabase(&CLISession.Config)
	if err != nil || db == nil {
		return errors.New("Could not create the database session : " + err.Error())
	}

	deldb := prompt("dbdel")
	willDeleteDB, err := yesNo(deldb)
	if willDeleteDB {
		err = db.DeleteDatabase()
		if err != nil {
			return errors.New("Could not delete the database : " + err.Error())
		}
	}

	err = db.SetupAuth(
		CLISession.Config.Database.User,
		CLISession.Config.Database.Password,
		CLISession.Config.Database.Database,
	)

	if err != nil {
		return errors.New("Could not setup the auth for the database : " + err.Error())
	}

	err = db.SetupDatabase()
	if err != nil {
		return errors.New("Could not setup the database : " + err.Error())
	}

	return nil
}

func setupAPI() error {
	type user struct {
		User     string
		Password string
	}

	p, err := strconv.Atoi(prompt("apiport"))
	if err != nil {
		return err
	}
	CLISession.Config.API.Port = uint16(p)

	wantToCreateUser := prompt("createuser")
	createBool, err := yesNo(wantToCreateUser)

	if createBool {
		CLISession.Config.API.User = prompt("apiuser")
		password := prompt("apipassword")
		db, err := database.NewDatabase(&CLISession.Config)
		if err != nil || db == nil {
			return errors.New("Could not create the database session : " + err.Error())
		}
		now := time.Now()

		hashedPassword, err := security.HashAndSalt([]byte(password))
		if err != nil {
			return err
		}
		user := models.User{Name: CLISession.Config.API.User, Password: hashedPassword, CreatedAt: now, UpdatedAt: now}

		err = db.CreateOrUpdateUser(user)
		if err != nil {
			return errors.New("Could not create user : " + err.Error())
		}
		log.Debugf("user : %+v", user)
		err = db.GenerateNewToken(user)
		if err != nil {
			return errors.New("Could not generate new token for user : " + err.Error())
		}

		user, err = db.GetUser(user.Name)
		if err != nil {
			return errors.New("Could not get newly created user : " + err.Error())
		}

		CLISession.Config.API.Token = user.Token
	}

	return nil
}

func setupServer() error {
	CLISession.Config.Server.IP = prompt("serverip")
	p, err := strconv.Atoi(prompt("serverport"))
	if err != nil {
		return err
	}

	CLISession.Config.Server.Port = uint16(p)
	CLISession.Config.Server.Password = prompt("serverpassword")

	//loop until answer is 'y' or 'n'
	tlsString := prompt("usetls")
	tlsBool, err := yesNo(tlsString)
	for err != nil {
		tlsString := prompt("usetls")
		tlsBool, err = yesNo(tlsString)
	}

	CLISession.Config.TLSParams.UseTLS = tlsBool
	return nil
}

func setupAlgorithm() error {
	var err error
	usedAlgo := prompt("algorithm")
	CLISession.Config.Algorithm.Name = usedAlgo

	return err
}

func setupWordlists() error {

	// passwords
	index := printWordlistByType("passwords")
	askPasswordWordlist := prompt("wordlist-password")
	selectedPasswordIndex, err := strconv.Atoi(askPasswordWordlist)
	if err != nil {
		return errors.New("Your input is not a number")
	}

	if selectedPasswordIndex < 0 || selectedPasswordIndex >= index {
		return errors.New("Input out of range")
	}

	// subdomains
	index = printWordlistByType("subdomains")
	askDomainWordlist := prompt("wordlist-domain")
	selectedDomainIndex, err := strconv.Atoi(askDomainWordlist)
	if err != nil {
		return errors.New("Your input is not a number")
	}

	if selectedDomainIndex < 0 || selectedDomainIndex >= index {
		return errors.New("Input out of range")
	}

	// return the selected "wordlist" struct from its index number
	// This is not optimal. (TOFIX)
	passwordWordlist, err := getWordlistByTypeAndIndex("passwords", selectedPasswordIndex)
	domainWordlist, err := getWordlistByTypeAndIndex("subdomains", selectedDomainIndex)

	fmt.Printf("Downloading : [%s] and [%s] wordlists\n", passwordWordlist.Name, domainWordlist.Name)
	err = downloadWordlist(passwordWordlist)
	if err != nil {
		return errors.New("Could not download the password wordlist : " + err.Error())
	}

	err = downloadWordlist(domainWordlist)
	if err != nil {
		return errors.New("Could not download the subdomains wordlist : " + err.Error())
	}

	fmt.Println("Download complete")
	return nil
}

func yesNo(response string) (bool, error) {
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		return true, nil
	} else if response == "n" || response == "no" {
		return false, nil
	}
	return false, errors.New("Invalid input")
}

func loadExistingConfig() (config.ConfigToml, error) {
	var data config.ConfigToml
	_, err := toml.DecodeFile(configPath, &data)

	if err != nil {
		return config.ConfigToml{}, errors.New("Couldn't decode config files : " + err.Error())
	}

	return data, nil
}

// Save the config on disk
func saveConfigFile() error {

	//Create new config file
	file, err := os.OpenFile(
		CLISession.ConfigPath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)

	if err != nil {
		return err
	}

	defer file.Close()

	//Write new config to new config file
	err = toml.NewEncoder(file).Encode(CLISession.Config)
	if err != nil {
		return err
	}

	return nil
}

//GeneratePassword returns a new password of len n (1st arg)
func GeneratePassword(n int) (string, error) {
	const codeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789&#{}[]()-|_^@=+%?./!;,"
	pass := ""
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for _, char := range b {
		pos := uint8(char)
		pass += string(codeAlphabet[int(pos)%len(codeAlphabet)])
	}

	return pass, nil
}

func init() {

	rootCmd.AddCommand(setupCmd)

	setupCmd.AddCommand(setupServerCmd)
	setupCmd.AddCommand(setupDatabaseCmd)
	setupCmd.AddCommand(setupAPICmd)
	setupCmd.AddCommand(setupTLSCmd)
	setupCmd.AddCommand(setupAlgorithmCmd)
	setupCmd.AddCommand(setupWordlistsCmd)

	// database
	setupCmd.PersistentFlags().StringVar(&cliDBSetupUser, "database-user", defaultDBSetupUser, "Custom database user")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupPassword, "database-password", defaultDBSetupPassword, "Custom database password")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupIP, "database-ip", defaultDBIP, "Custom database ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliDBSetupPort, "database-port", defaultDBPort, "Custom database port number")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupType, "database-type", defaultDBType, "Custom database type number")

	//server
	setupCmd.PersistentFlags().StringVar(&cliServerPassword, "server-password", defaultServerPassword, "Custom server password")
	setupCmd.PersistentFlags().StringVar(&cliServerIP, "server-ip", defaultServerIP, "Custom server ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliServerPort, "server-port", defaultServerPort, "Custom server port number")

	//api
	setupCmd.PersistentFlags().StringVar(&cliApiUser, "api-user", defaultAPIUser, "Custom API user")
	setupCmd.PersistentFlags().StringVar(&cliApiPassword, "api-password", defaultAPIPassword, "Custom API password")
	setupCmd.PersistentFlags().StringVar(&cliApiIP, "api-ip", defaultAPIIP, "Custom API ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliApiPort, "api-port", defaultAPIPort, "Custom API port number")

	//TLS
	setupCmd.PersistentFlags().BoolVar(&cliDisableTLS, "disable-tls", false, "Disable TLS")

	//Algorithm
	setupCmd.PersistentFlags().StringVar(&cliAlgorithm, "algorithm", defaultAlgorithm, "Load balancing algorithm")

	//Wordlist
	setupCmd.PersistentFlags().StringVar(&cliWordlist, "wordlists", defaultAlgorithm, "Download wordlists (password, subdomains...)")

	//Skips
	setupCmd.PersistentFlags().BoolVar(&skipDBSetup, "skip-database", false, "Skip configuration of the database")
	setupCmd.PersistentFlags().BoolVar(&skipServerSetup, "skip-server", false, "Skip configuration of the server")
	setupCmd.PersistentFlags().BoolVar(&skipApiSetup, "skip-api", false, "Skip configuration of the Api")
	setupCmd.PersistentFlags().BoolVar(&skipAlgorithmSetup, "skip-algorithm", false, "Skip configuration of the algorithm")
	setupCmd.PersistentFlags().BoolVar(&skipModulesSetup, "skip-modules", false, "Skip configuration of the modules")
	setupCmd.PersistentFlags().BoolVar(&skipProjectSetup, "skip-project", false, "Skip configuration of the project")
	setupCmd.PersistentFlags().BoolVar(&skipWordlists, "skip-wordlists", false, "Skip downloading of wordlists")
}
