package server

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/netm4ul/netm4ul/cmd/colors"
	"io/ioutil"
	"log"
)

// Build the TLS configuration for server
func TLSBuildServerConf() *tls.Config {

	caCert := "certificates/NetM4ul_CA.crt"
	publKey := "certificates/NetM4ul_Server.crt"
	privKey := "certificates/NetM4ul_Server.pem"

	// Read CA file and initialise
	caCertBytes, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Println(colors.Red("Unable to read %s : %s"), caCert, err.Error())
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		log.Println(colors.Red("Unable to add CA certificate to certificate pool"))
	}

	// Read own KeyPair
	cert, err := tls.LoadX509KeyPair(publKey, privKey)
	if err != nil {
		log.Println(err)
		log.Println(colors.Red("Unable to read X509KeyPair : %s"), err.Error())
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

	return tlsConfig
}
