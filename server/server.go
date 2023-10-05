//
// Copyright (c) 2023 Tenebris Technologies Inc.
// See LICENSE for further information.
//

package server

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"golang.org/x/net/netutil"
	"log"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Listen          string
	HTTPTimeout     int
	HTTPIdleTimeout int
	MaxConcurrent   int
	TLSCertFile     string
	TLSKeyFile      string
	Debug           bool
	Logger          *log.Logger
	server          *http.Server
}

// New returns a new Config struct with default values
func New() Config {

	// Return default configuration
	return Config{
		Listen:          "127.0.0.1:443",
		HTTPTimeout:     5,
		HTTPIdleTimeout: 5,
		MaxConcurrent:   100,
		TLSCertFile:     "cert.pem",
		TLSKeyFile:      "key.pem",
		Debug:           false,
		Logger:          nil,
	}
}

// Start starts the API
func (c *Config) Start() error {

	// Create server with a single handler (no routing is required)
	s := &http.Server{
		Addr:              c.Listen,
		Handler:           c.Wrapper(http.HandlerFunc(c.Handler)),
		ReadHeaderTimeout: time.Duration(c.HTTPTimeout) * time.Second,
		ReadTimeout:       time.Duration(c.HTTPTimeout) * time.Second,
		WriteTimeout:      time.Duration(c.HTTPTimeout) * time.Second,
		IdleTimeout:       time.Duration(c.HTTPIdleTimeout) * time.Second,
		ErrorLog:          c.Logger,
	}

	if c.TLSCertFile == "" || c.TLSKeyFile == "" {
		return errors.New("TLS cert or key file not specified")
	}

	// Load the cert and key
	cert, err := tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
	if err != nil {
		return err
	}

	// Log certificate information
	c.logHash(sha256.Sum256(cert.Certificate[0]))
	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err == nil {
		c.Logger.Printf("Certificate CN %s valid from %+v to %+v",
			parsed.Subject.CommonName, parsed.NotBefore, parsed.NotAfter)
	}

	// Create the TLS configuration
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}

	// Set minimum TLS version to 1.2 to avoid vulnerability scanners complaining
	tlsConfig.MinVersion = tls.VersionTLS12

	// Add to the HTTP server config
	s.TLSConfig = &tlsConfig

	// Start our customized server
	return c.listen(s)
}

func (c *Config) Stop() error {

	// Tell the server it has 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Protect against nil server
	if c.server == nil {
		return errors.New("server is not running")
	}

	// Shutdown the server
	if err := c.server.Shutdown(ctx); err != nil {
		return errors.New(fmt.Sprintf("server shutdown error: %s", err.Error()))
	}

	// Shutdown was successful
	return nil
}

// listen is a replacement for ListenAndServe that implements a concurrent session limit
// using netutil.LimitListener. If maxConcurrent is 0, no limit is imposed.
func (c *Config) listen(srv *http.Server) error {

	// Store the server to allow for a graceful shutdown
	c.server = srv

	// Get listen address, default to ":http"
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}

	// Create listener
	rawListener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// If maxConcurrent > 0 wrap the listener with a limited listener
	var listener net.Listener
	if c.MaxConcurrent > 0 {
		listener = netutil.LimitListener(rawListener, c.MaxConcurrent)
	} else {
		listener = rawListener
	}

	// This will use the previously configured TLS information
	c.Logger.Printf("Starting HTTPS server on %s", addr)
	return srv.ServeTLS(listener, "", "")
}

// Handler implements the http.Handler interface
func (c *Config) Handler(w http.ResponseWriter, _ *http.Request) {

	// Set reply headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "-1")

	// Send reply with HTTP 200
	w.WriteHeader(http.StatusOK)
}

// Log the certificate hash
func (c *Config) logHash(hash [32]byte) {
	var s string
	for _, f := range hash {
		s += fmt.Sprintf("%02X", f)
	}
	c.Logger.Printf("TLS certificate SHA256: %s", s)
}
