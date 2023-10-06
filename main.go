//
// Copyright (c) 2023 Tenebris Technologies Inc.
// All rights reserved.
//

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ZTBeacon/cert"
	"ZTBeacon/global"
	"ZTBeacon/server"
	"ZTBeacon/slog"
)

var logger *log.Logger

func main() {
	var err error

	// Configure command line arguments
	newCert := flag.Bool("newcert", false, "generate a new self-signed certificate and key")
	certFile := flag.String("cert", "cert.pem", "CA certificate file in PEM format")
	keyFile := flag.String("key", "key.pem", "CA key file in PEM format")
	logFile := flag.String("log", "ztbeacon.log", "log file")
	debug := flag.Bool("debug", false, "enable debug logs")
	console := flag.Bool("console", false, "force logging to stdout")
	listen := flag.String("listen", "0.0.0.0:4433", "listen address and port")
	deny := flag.String("deny", "", "deny IP address access to the server")
	flag.Parse()

	// Create a logger using the slog package
	logger, err = slog.New("ztbeacon", *logFile, *console, *debug)
	if err != nil {
		fmt.Printf("Error creating logger: %s", err)
		os.Exit(1)
	}

	logger.Printf("%s v%s starting", global.ProductName, global.ProductVersion)

	// Setup signal catching
	signals := make(chan os.Signal, 1)

	// Catch signals
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Method invoked upon receiving signal
	go func() {
		for {
			s := <-signals
			logger.Printf("received signal: %v", s)

			// Graceful exit
			cleanup()
		}
	}()

	// If newcert is specified, generate a self-signed certificate
	if *newCert {
		logger.Printf("Generating self-signed certificate")
		err := cert.New(*certFile, *keyFile)
		if err != nil {
			logger.Println("Error generating self-signed certificate:", err)
			os.Exit(1)
		}
		logger.Printf("Self-signed certificate generated and saved to %s %s", *certFile, *keyFile)
	}

	// Set server parameters
	s := server.New()
	s.Listen = *listen
	s.HTTPTimeout = 5
	s.HTTPIdleTimeout = 5
	s.MaxConcurrent = 100
	s.Debug = *debug
	s.TLSCertFile = *certFile
	s.TLSKeyFile = *keyFile
	s.Logger = logger

	if *deny != "" {
		s.AddDeny(*deny)
	}

	// Start the server
	err = s.Start()
	if err != nil {
		// Server returns an error even if it shut down gracefully
		// If the error is not a server closed error, return it
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("Error starting the HTTPS server: %s", err.Error())
		} else {
			logger.Printf("HTTPS server shut down gracefully")
		}
	}
	cleanup()
}

// cleanup is the graceful exit point
func cleanup() {

	// Perform any cleanup here

	// Exit
	logger.Printf("%s %s exiting", global.ProductName, global.ProductVersion)
	os.Exit(0)
}
