package server

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"crypto/x509"
)

// Build the TLS configuration for server
func TLSBuildServerConf() (tls.Config){

	// Read CA file and initialise
	certBytes, err := ioutil.ReadFile("certificates/NetM4ul_CA.crt")
	if err != nil {
		log.Fatalln("Unable to read localhost.pem", err)
	}

	clientCertPool := x509.NewCertPool()
	if ok := clientCertPool.AppendCertsFromPEM(certBytes); !ok {
		log.Fatalln("Unable to add certificate to certificate pool")
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
	}

	return tlsConfig
}