//
// Copyright (c) 2023 Tenebris Technologies Inc.
// All rights reserved.
//

package cert

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"ZTBeacon/global"
)

func New(certFileName string, keyFileName string) error {

	// Create a serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to generate serial number: %s", err.Error()))
	}

	// Set up certificate. We'll make this both a CA and server certificate to cover the bases.
	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         global.ProductName,
			Locality:           []string{global.ProductName},
			Organization:       []string{global.ProductName + " Self Signed"},
			OrganizationalUnit: []string{global.ProductName + " Self Signed"},
			Country:            []string{"XX"},
		},
		Issuer: pkix.Name{
			CommonName:         global.ProductName + " Self Signed",
			Locality:           []string{global.ProductName},
			Organization:       []string{global.ProductName + " Self Signed"},
			OrganizationalUnit: []string{global.ProductName + " Self Signed"},
			Country:            []string{"XX"},
		},
		DNSNames:    []string{global.ProductName, global.ProductName},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(100, 0, 0),
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDataEncipherment |
			x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	// Create our private and public key
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Create the CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privKey.PublicKey, privKey)
	if err != nil {
		return err
	}

	// Write a PEM encoded CA certificate file
	certFile, err := os.Create(certFileName)
	if err != nil {
		return err
	}

	certWriter := bufio.NewWriter(certFile)

	err = pem.Encode(certWriter, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return err
	}

	_ = certWriter.Flush()
	_ = certFile.Close()

	// Write a PEM encoded CA private key file
	keyFile, err := os.Create(keyFileName)
	if err != nil {
		return err
	}

	keyWriter := bufio.NewWriter(keyFile)

	err = pem.Encode(keyWriter, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	if err != nil {
		return err
	}

	_ = keyWriter.Flush()
	_ = keyFile.Close()
	return nil
}
