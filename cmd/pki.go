// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/spf13/cobra"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/spf13/viper"
	"path/filepath"
	"strconv"
)

var (
	nNodes               uint
	serverID             string
	ecdsaCurve           string
	keepPrivCA           bool
	certDuration         time.Duration
	organisationSubject  string
	certificateDirectory string
)

// pkiCmd represents the pki command
var pkiCmd = &cobra.Command{
	Use:   "pki",
	Short: "Set up PKI (CA, Server and clients)",
	Long: `This command builds a Public Key Infrastructure (PKI) for mutual TLS authentication between your server and the nodes.
	A Certificate Authority (CA) is created, signing the server (master node) certificate, as well as all the client nodes at first deploy time.
	The CA certificate is shipped to server and clients along with own certificate and private key of an entity, thus enabling mutual authentication on TLS handshake.
	
	The parameters are to be specified in the configuration file or inline.
	
	The master node has probably an IP address as hostname, if you run your instance from a box without DNS indexing.
	
	The client nodes don't need a publicly known hostname, but you may assign them a name that suits your use (e.g. your logs may see "client_x successfully connected to API"")`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.LoadConfig(configPath)

		if serverID == "" {
			serverID = config.Config.Server.IP
		}

	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setting up PKI at " + certificateDirectory)
		pkiSetup(organisationSubject, nNodes, certDuration, ecdsaCurve, certificateDirectory, serverID)
	},
}

func init() {
	rootCmd.AddCommand(pkiCmd)

	pkiCmd.Flags().UintVarP(&nNodes, "nodes", "n", 1, "Number of nodes to create certificates for")
	pkiCmd.Flags().StringVarP(&serverID, "server", "@", "", "Server address, with IP or DNS name")
	pkiCmd.Flags().StringVarP(&ecdsaCurve, "ec", "e", "P256", "Elliptic curve to use for certificates. Accepted values : P224, P256 (recommended), P384 and P521")
	pkiCmd.Flags().BoolVarP(&keepPrivCA, "keepca", "k", false, "Indicate whether to keep CA private key (i.e. written to disk)")
	pkiCmd.Flags().DurationVarP(&certDuration, "duration", "d", 365*24*time.Hour, "Duration for certificate. Default is a year")
	pkiCmd.Flags().StringVarP(&organisationSubject, "org", "o", "Netm4ul", "Organisation Subject to use in server certificate")
	pkiCmd.Flags().StringVarP(&certificateDirectory, "dir", "t", "./certificates", "Local directory to store PKI certs and keys")

	viper.SetDefault("ec", "P256")
	viper.SetDefault("keepca", false)
	viper.SetDefault("duration", 365*24*time.Hour)
	viper.SetDefault("subject", "Netm4ul")
	viper.SetDefault("dir", "./certificates")

	pkiCmd.MarkFlagRequired("nodes")
}

// Extract public key from private
func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	/*case *rsa.PrivateKey:
	return &k.PublicKey*/
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

/*
func get_formatted_time() string {
	now := time.Now()

	nowH, nowM, nowS := now.Clock()

	return fmt.Sprintf("%s %d %d:%d:%d %d", now.Month(), now.Day(), nowH, nowM, nowS, now.Year())
}
*/

// Generates a ecdsa private key from a cryptographically secure prng
func generatePrivateKey(ecdsaCurve string) interface{} {
	var priv interface{}
	var err error
	switch ecdsaCurve {
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized elliptic curve: %q", ecdsaCurve)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	return priv
}

// Geneates a random serial number
func generateSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	return serialNumber
}

// Wrapper function to create a certificate from private key and certificate template
func buildCert(template *x509.Certificate, signerCert *x509.Certificate, privateKey interface{}, signerPrivKey interface{}) []byte {
	derBytes, err := x509.CreateCertificate(rand.Reader, template, signerCert, publicKey(privateKey), signerPrivKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	return derBytes
}

// Wrapper function to generate the CA certificate
func buildCACert(template *x509.Certificate, privateKey interface{}) ([]byte, *x509.Certificate) {
	template.IsCA = true
	template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign

	derBytes := buildCert(template, template, privateKey, privateKey)

	caCert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		log.Fatalf("Failed to parse CA certificate: %s", err)
	}

	return derBytes, caCert
}

// Wrapper function to generate the server certificate
func buildServerCert(template *x509.Certificate, host string, privateKey interface{}, signerCert *x509.Certificate, signerPrivKey interface{}) []byte {

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}

	return buildCert(template, signerCert, privateKey, signerPrivKey)
}

// Wrapper function to generate the client certificate
func buildClientCert(template *x509.Certificate, id string, privateKey interface{}, signerCert *x509.Certificate, signerPrivKey interface{}) []byte {
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	template.Subject.CommonName = id

	return buildCert(template, signerCert, privateKey, signerPrivKey)
}

// Write key material to disk. Note : the CA's private key is not written to disk.
func writeCertAndKeyToDisk(entityType string, privateKey interface{}, certFilename string, keyFilename string, derBytes []byte) {

	// Make sure target directory exists
	err := os.MkdirAll(filepath.Dir(certFilename), 0700)
	if err != nil {
		log.Fatalf("failed to create directory for cert : %s.\n%s", certFilename, err)
	}

	// Write Certificate to File
	certOut, err := os.Create(certFilename)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	// log.Print("wrote " + certFilename + "\n")

	// Write key to file if it is not a CA
	if entityType != "CA" || keepPrivCA {
		keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Print("failed to open "+keyFilename+"  for writing:", err)
		}
		pem.Encode(keyOut, pemBlockForKey(privateKey))
		keyOut.Close()
		// log.Print("wrote " + keyFilename + "\n")
		// log.Print("Successfully achieved job for " + keyFilename + "\n")
	}

}

// Creates the certificate for the given entity and writes key material to disk. In case of the CA, also return the certificate and private key to sign the end entity parties
func create(entityID string, validFor time.Duration, entityType string, ecdsaCurve string, signerCert *x509.Certificate, signerPrivKey interface{}, organisation string, destDir string) (*x509.Certificate, interface{}, string, string) {

	if len(entityID) == 0 {
		log.Fatalf("Missing required --host parameter")
	}

	// Generate private key
	privateKey := generatePrivateKey(ecdsaCurve)

	// Define timespan the certificate has to be valid for
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	// Generate a random serial number
	serialNumber := generateSerial()

	// Define a name for the certificate
	certFilename := destDir + "/" + entityID + "/" + entityID + "_cert.pem"
	keyFilename := destDir + "/" + entityID + "/" + entityID + "_key.pem"

	// Build Certificate Template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organisation},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	// Conditional additions based on entity type : CA, server or client
	// Create key and certificate in DER format
	var derBytes []byte
	var caCert *x509.Certificate

	switch entityType {

	case "CA":
		derBytes, caCert = buildCACert(&template, privateKey)

	case "Server":
		template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		derBytes = buildServerCert(&template, entityID, privateKey, signerCert, signerPrivKey)

	case "Client":
		template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		derBytes = buildClientCert(&template, entityID, privateKey, signerCert, signerPrivKey)

	default:
		log.Fatalf("Certificate entity type not recognised : %s", "var")

	}

	// Write certificate to file
	writeCertAndKeyToDisk(entityType, privateKey, certFilename, keyFilename, derBytes)

	if entityType == "CA" {
		return caCert, privateKey, certFilename, keyFilename
	}

	return nil, nil, certFilename, keyFilename
}

// Checks if directory exists. If not, creates it.
func checkDir(targetDirectory string) {

	//newpath := filepath.Join(".", targetDirectory)
	err := os.MkdirAll(targetDirectory, 0700)
	if err == nil {
		log.Println("WARNING : The certificate directory already exists. All certificates and keys that already exist will be overwritten if they have the same name as the new. The old, overwritten data would not be usable nor recoverable.")
	}
}

func pkiSetup(organisationSubject string, numberClients uint, certDuration time.Duration, ecdsaCurve string, certificateDirectory string, serverID string) {

	// Perform Checks and pre-setup
	checkDir(certificateDirectory)

	// Create CA
	caCert, caKey, _, _ := create("ca", certDuration, "CA", ecdsaCurve, nil, nil, organisationSubject, certificateDirectory)
	if caCert == nil || caKey == nil {
		log.Print("Failed to created CA cert\n")
	}

	// Create Server
	_, _, _, _ = create(serverID, certDuration, "Server", ecdsaCurve, caCert, caKey, organisationSubject, certificateDirectory)

	// Create clients
	clientsDir := certificateDirectory + "/clients"
	for i := 0; i < int(numberClients); i++ {
		_, _, _, _ = create("client_"+strconv.Itoa(i), certDuration, "Client", ecdsaCurve, caCert, caKey, organisationSubject, clientsDir)

	}

	log.Println("It seems PKI creation was successfull.")

}
