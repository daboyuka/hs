package main

import (
	"crypto/tls"
	_ "embed"
)

//go:embed server.pem
var pemData []byte

//go:embed server.key
var keyData []byte

func init() {
	cert, err := tls.X509KeyPair(pemData, keyData)
	if err != nil {
		panic(err)
	}
	server.TLSConfig.Certificates = append(server.TLSConfig.Certificates, cert)
}
