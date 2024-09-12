/*
Package tools provides utilities for tests.

certsetup() is one of those utilities.
Full credit for it goes to https://gist.github.com/shaneutt

This code was pulled and modified from the following resources:
- https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
- https://shaneutt.com/blog/golang-ca-and-signed-cert-go/.

USAGE:

	go run generate_cert.go

This will generate a CA cert/key pair and use that to sign Server cert/key pair
and Client cert/key pair.

Use these certs for tests such as websupport_test and orchestratorNoopProvider.
*/
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	// get our ca and server certificate
	err := certsetup()
	if err != nil {
		panic(err)
	}
}

func certsetup() (err error) {
	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"Strata Identity"},
			Country:      []string{"US"},
			Province:     []string{"CO"},
			Locality:     []string{"Boulder"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	log.Println("Writing out CA Cert")
	err = os.WriteFile("ca-cert.pem", caPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	log.Println("Writing out CA Key")
	err = os.WriteFile("ca-key.pem", caPrivKeyPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	// set up our server certificate
	certPEM, certPrivKeyPEM, err := generateCert(
		ca,
		caPrivKey,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return err
	}
	log.Println("Writing out Server Cert")
	err = os.WriteFile("server-cert.pem", certPEM, 0644)
	if err != nil {
		return err
	}
	log.Println("Writing out Server Key")
	err = os.WriteFile("server-key.pem", certPrivKeyPEM, 0644)
	if err != nil {
		return err
	}

	// set up our client certificate
	clientCertPEM, clientCertPrivKeyPEM, err := generateCert(
		ca,
		caPrivKey,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	)
	if err != nil {
		return err
	}

	log.Println("Writing out Client Cert")
	err = os.WriteFile("client-cert.pem", clientCertPEM, 0644)
	if err != nil {
		return err
	}
	log.Println("Writing out Client Key")
	err = os.WriteFile("client-key.pem", clientCertPrivKeyPEM, 0644)
	if err != nil {
		return err
	}

	return
}

func generateCert(
	ca *x509.Certificate,
	caPrivKey *rsa.PrivateKey,
	keyUsage []x509.ExtKeyUsage,
) ([]byte, []byte, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"Strata Identity"},
			Country:      []string{"US"},
			Province:     []string{"CO"},
			Locality:     []string{"Boulder"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  keyUsage,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}
