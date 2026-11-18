package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

func tlsClient(certFile, keyFile, caFile string) (*http.Client, error) {
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}}}, nil
}

func serverTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool, MinVersion: tls.VersionTLS12}, nil
}

func tlsServer(addr string, h http.Handler, certFile, keyFile, caFile string) (*http.Server, error) {
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("bad ca")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:    addr,
		Handler: h,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    pool,
			MinVersion:   tls.VersionTLS12,
		},
	}, nil
}
