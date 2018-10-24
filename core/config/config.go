package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// API : Rest API config
type API struct {
	IP    string `toml:"ip"`
	Port  uint16 `toml:"port"`
	User  string `toml:"user"`
	Token string `toml:"token"`
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

// Database : Mongodb config
type Database struct {
	User         string `toml:"user"`
	Database     string `toml:"database"`
	DatabaseType string `toml:"databaseType"`
	Password     string `toml:"password"`
	IP           string `toml:"ip"`
	Port         uint16 `toml:"port"`
}

// Module : Basic struct for general module config
type Module struct {
	Enabled bool `toml:"enabled" json:"enabled"`
}

// // Versions : Store the version
// type Versions struct {
// 	Api    string `toml:"api" json:"api"`
// 	Server string `toml:"server" json:"server"`
// 	Client string `toml:"client" json:"client"`
// }

type Project struct {
	Name        string `toml:"name" json:"name"`
	Description string `toml:"description" json:"description"`
}

type Algorithm struct {
	Name string `toml:"name"`
	Mode string
}

// ConfigToml is the global config object
type ConfigToml struct {
	Project Project
	// Versions  Versions
	API       API
	DNS       DNS
	Keys      Keys
	Server    Server
	Database  Database
	Modules   map[string]Module
	TLSParams TLSParams
	Algorithm Algorithm
}

// LoadConfig load the configuration file !
func LoadConfig(file string) (ConfigToml, error) {
	var configPath string
	var config ConfigToml

	if file == "" {
		dir, err := os.Getwd()

		if err != nil {
			return config, err
		}
		configPath = filepath.Join(dir, "netm4ul.conf")
	} else {
		configPath = file
	}

	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// Read CA file and initialise
func (c *ConfigToml) TLSReadCAFile(caCert string) (*x509.CertPool, error) {

	caCertBytes, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Errorf("Unable to read CA file %s : %s", caCert, err.Error())
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		log.Error("Unable to add CA certificate to certificate pool")
		return nil, err
	}

	return caCertPool, nil
}

// Build the TLS configuration for server
func (c *ConfigToml) TLSBuildServerConf() (*tls.Config, error) {

	// Get CA file
	caCertPool, err := c.TLSReadCAFile(c.TLSParams.CaCert)
	if err != nil {
		return nil, err
	}

	// Read own KeyPair
	cert, err := tls.LoadX509KeyPair(c.TLSParams.ServerCert, c.TLSParams.ServerPrivateKey)
	if err != nil {
		log.Errorf("Unable to read X509KeyPair at %s : %s", c.TLSParams.ServerCert, err.Error())
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
func (c *ConfigToml) TLSBuildClientConf() (*tls.Config, error) {

	// Read CA file and initialise
	caCertPool, err := c.TLSReadCAFile(c.TLSParams.CaCert)
	if err != nil {
		return nil, err
	}

	// Read own KeyPair
	cert, err := tls.LoadX509KeyPair(c.TLSParams.ClientCert, c.TLSParams.ClientPrivateKey)
	if err != nil {
		log.Errorf("Unable to read Client X509KeyPair : %s", err.Error())
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
