package config

import (
	"log"
	"os"
	"path/filepath"

	"crypto/tls"
	"crypto/x509"
	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/cmd/colors"
	"io/ioutil"
	"net"
)

// API : Rest API config
type API struct {
	Port     uint16 `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// DNS : Setup DNS resolver IP
type DNS struct {
	Resolvers string `toml:"resolvers"`
}

// Keys : setup tocken & api keys
type Keys struct {
	Google string `toml:"google"`
	Shodan string `toml:"shodan"`
}

// Server : Master node config
type Server struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
	IP       string `toml:"ip"`
	Port     uint16 `toml:"port"`
}

// TLS Certificates and keys
type TLSParams struct {

	// Enable TLS communications or not
	UseTLS bool `toml:"usetls"`

	// TLS Configuration to be set later
	TLSConfig *tls.Config

	// Certifcation Authority and Server Certificates
	CaCert     string `toml:"caCert"`
	ServerCert string `toml:"serverCert"`

	// These will later be deleted to be load dynamically for server and clients
	ServerPrivateKey string `toml:"serverPrivateKey"`
	ClientCert       string `toml:"clientCert"`
	ClientPrivateKey string `toml:"clientPrivateKey"`
}

// Connection type, to handle either use of TLS or not
type Connector struct {
	TLSConn *tls.Conn
	Conn    net.Conn
}

// Database : Mongodb config
type Database struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
	IP       string `toml:"ip"`
	Port     uint16 `toml:"port"`
}

// Module : Basic struct for general module config
type Module struct {
	Enabled bool `toml:"enabled" json:"enabled"`
}

// Versions : Store the version
type Versions struct {
	Api    string `toml:"api" json:"api"`
	Server string `toml:"server" json:"server"`
	Client string `toml:"client" json:"client"`
}

// Node : Node info
type Node struct {
	Modules []string `json:"modules"`
	Project string   `json:"project"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ConfigToml is the global config object
type ConfigToml struct {
	Project    Project
	Versions   Versions
	Verbose    bool
	NoColors   bool
	ConfigPath string
	Mode       string
	IsServer   bool
	IsClient   bool
	Targets    []string
	API        API
	DNS        DNS
	Keys       Keys
	Server     Server
	Database   Database
	Nodes      map[string]Node
	Modules    map[string]Module
	Connector  Connector
	TLSParams  TLSParams
}

// Config : exported config
var Config ConfigToml

// LoadConfig load the configuration file !
func LoadConfig(file string) {
	var configPath string

	if file == "" {
		dir, err := os.Getwd()

		if err != nil {
			log.Fatal(err)
		}

		configPath = filepath.Join(dir, "netm4ul.conf")
	} else {
		configPath = file
	}

	if _, err := toml.DecodeFile(configPath, &Config); err != nil {
		log.Fatalln(err)
	}
}

// Read CA file and initialise
func TLSReadCAFile(caCert string) (*x509.CertPool, error) {

	caCertBytes, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Printf(colors.Red("Unable to read CA file %s : %s"), caCert, err.Error())
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		log.Println(colors.Red("Unable to add CA certificate to certificate pool"))
		return nil, err
	}

	return caCertPool, nil
}

// Build the TLS configuration for server
func TLSBuildServerConf() (*tls.Config, error) {

	var err error
	var caCertPool *x509.CertPool

	// Get CA file
	caCertPool, err = TLSReadCAFile(Config.TLSParams.CaCert)
	if err != nil {
		return nil, err
	}

	// Read own KeyPair
	cert, err := tls.LoadX509KeyPair(Config.TLSParams.ServerCert, Config.TLSParams.ServerPrivateKey)
	if err != nil {
		log.Printf(colors.Red("Unable to read X509KeyPair at %s : %s"), Config.TLSParams.ServerCert, err.Error())
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,
		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: caCertPool,
		// PFS because we can
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		// Force it server side
		PreferServerCipherSuites: true,
		// TLS 1.2 because we can
		MinVersion: tls.VersionTLS12,
		// Server Key Material
		Certificates: []tls.Certificate{cert},
		// Set preferences for used curves (but certs should already have been made)
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
	}

	return tlsConfig, nil
}

// Build the TLS configuration for server
func TLSBuildClientConf() (*tls.Config, error) {

	var err error
	var caCertPool *x509.CertPool

	// Read CA file and initialise
	caCertPool, err = TLSReadCAFile(Config.TLSParams.CaCert)
	if err != nil {
		return nil, err
	}

	// Read own KeyPair
	cert, err := tls.LoadX509KeyPair(Config.TLSParams.ClientCert, Config.TLSParams.ClientPrivateKey)
	if err != nil {
		log.Printf(colors.Red("Unable to read Client X509KeyPair : %s"), err.Error())
		return nil, err
	}

	tlsConfig := &tls.Config{
		// Ensure that we only use our "CA" to validate certificates
		RootCAs: caCertPool,
		// Server Key Material
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}
