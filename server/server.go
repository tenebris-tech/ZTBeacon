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
	"strings"
	"sync"
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
	denyList        denyList
}

type denyList struct {
	ipList []string
	mu     sync.RWMutex
}

// Create a log writer that throws away all data
type nullWriter struct {
}

func (n *nullWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
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
	}

	if c.Debug {
		s.ErrorLog = c.Logger
	} else {
		// Suppress server error logging because it is noisy and shows every dropped connection
		s.ErrorLog = log.New(&nullWriter{}, "", 0)
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

	// Wrap the listener so that we can block selected connections
	return srv.ServeTLS(WrapListener(listener, c), "", "")
}

// WrapListener wraps a net.Listener and adds connection blocking in Accept()
// Pass the config so that various options are available
func WrapListener(orig net.Listener, config *Config) net.Listener {
	wrapper := Listener{
		orig:   orig,
		config: config,
	}
	return wrapper
}

// Listener implements a custom net.Listener interface to add logging and IP filtering
type Listener struct {
	orig   net.Listener
	config *Config
}

// Close closes the listener by calling the original listener
func (l Listener) Close() error {
	return l.orig.Close()
}

// Addr returns the address of the listener by calling the original listener
func (l Listener) Addr() net.Addr {
	return l.orig.Addr()
}

// Accept adds debug logging and IP filtering to the original listener
func (l Listener) Accept() (net.Conn, error) {

	// Accept the connection via the original listener
	conn, err := l.orig.Accept()

	// Get the remote IP
	remoteIP := conn.RemoteAddr().String()

	// Clean up, remove port number
	if len(remoteIP) > 0 {
		if strings.HasPrefix(remoteIP, "[") {
			// IPv6 address
			t := strings.Split(remoteIP, "]")
			remoteIP = t[0][1:]
		} else {
			// IPv4 - truncate :port
			t := strings.Split(remoteIP, ":")
			remoteIP = t[0]
		}
	}

	// Check if we should block this connection
	if l.config.deny(remoteIP) {
		if l.config.Debug {
			l.config.Logger.Printf("Connection from %s terminated", remoteIP)
		}

		// Close the connection. This will result in an error within the HTTP server.
		_ = conn.Close()
	} else {
		if l.config.Debug {
			l.config.Logger.Printf("Connection from %s allowed", remoteIP)
		}
	}

	return conn, err
}

// Handler implements the http.Handler interface
func (c *Config) Handler(w http.ResponseWriter, _ *http.Request) {

	// Set reply header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Send "OK" reply with HTTP 200
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "OK\n")
}

// Log the certificate hash
func (c *Config) logHash(hash [32]byte) {
	var s string
	for _, f := range hash {
		s += fmt.Sprintf("%02X", f)
	}
	c.Logger.Printf("TLS certificate SHA256: %s", s)
}
