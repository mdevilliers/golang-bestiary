package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {

	organisation := "acmecorp"

	// Common Name for the certificate. The common name
	// can be anything, but is usually set to the server's primary
	// DNS name. Even if you plan to connect via IP address you
	// should specify the DNS name here.
	commonName := "localhost"

	// additional DNS names and IP
	// addresses that clients may use to connect to the server. If
	// you plan to connect to the server via IP address and not DNS
	// then you must specify those IP addresses here.
	ipaddressOrDNSNames := []string{"127.0.0.1", "localhost"}

	// How long should the certificate be valid for? A year (365
	// days) is usual but requires the certificate to be regenerated
	// within a year or the certificate will cease working.
	validInDays := 365

	template := x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{organisation},
			CommonName:   commonName,
		},
		NotBefore: time.Now(),

		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		IsCA: true,
	}

	for _, val := range ipaddressOrDNSNames {
		if ip := net.ParseIP(val); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, val)
		}
	}

	template.NotAfter = template.NotBefore.Add(time.Duration(validInDays) * time.Hour * 24)

	spew.Dump(template)

	server_path := "../server"
	client_path := "../client"

	if _, err := os.Stat(server_path); os.IsNotExist(err) {
		os.Mkdir(server_path, 0700)
	}
	if _, err := os.Stat(client_path); os.IsNotExist(err) {
		os.Mkdir(client_path, 0700)
	}

	generateCerts(template, server_path)

	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}

	generateCerts(template, client_path)
}

func generateCerts(template x509.Certificate, path string) {

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		os.Exit(1)
	}

	fmt.Println("Private key :")
	spew.Dump(priv)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	template.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Println("Failed to generate serial number:", err)
		os.Exit(1)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Println("Failed to create certificate:", err)
		os.Exit(1)
	}

	fmt.Println("Cert bytes :")
	spew.Dump(derBytes)

	certOut, err := os.Create(path + "/selfsigned.crt")
	if err != nil {
		fmt.Println("Failed to open selfsigned.pem for writing:", err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(path+"/selfsigned.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("failed to open selfsigned.key for writing:", err)
		os.Exit(1)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
}
