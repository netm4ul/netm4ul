package server

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"crypto/x509"
	"github.com/netm4ul/netm4ul/cmd/colors"
)

// Build the TLS configuration for server
func TLSBuildServerConf() (tls.Config){

	caCert := "certificates/Netm4ul_CA.crt"
	publKey := "certificates/Netm4ul_Server.crt"
	privKey := "certificates/Netm4ul_Server.pem"

	// Read CA file and initialise
	certBytes, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Println(colors.Red("Unable to read %s : %s"), caCert, err.Error())
	}

	clientCertPool := x509.NewCertPool()
	if ok := clientCertPool.AppendCertsFromPEM(certBytes); !ok {
		log.Println(colors.Red("Unable to add CA certificate to certificate pool"))
	}

	// Read own KeyPair
	cert, err := tls.LoadX509KeyPair(publKey, privKey)
	if err != nil {
		log.Println(err)
		log.Println(colors.Red("Unable to read X509KeyPair : %s"), err.Error())
	}

	tlsConfig := tls.Config{
		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,
		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: clientCertPool,
		// PFS because we can
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		// Force it server side
		PreferServerCipherSuites: true,
		// TLS 1.2 because we can
		MinVersion: tls.VersionTLS12,
		// Server Key Material
		Certificates: []tls.Certificate{cert},
		// Set preferences for used curves (but certs should already have been made)
		CurvePreferences: []tls.CurveID{ tls.X25519, tls.CurveP256, tls.CurveP384},
	}

	return tlsConfig
}