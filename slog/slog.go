//
// Copyright (c) 2023 Tenebris Technologies Inc.
// All rights reserved.
//

package slog

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

// Map to allow multiple Loggers to use the same fileHandle
var logFiles = make(map[string]*os.File)

// New returns a new logger with the given name
func New(loggerName string, fileName string, console bool, debug bool) (*log.Logger, error) {
	var fileHandle *os.File = nil

	if fileName != "" {
		// Check to see if file handle already exists
		if _, ok := logFiles[fileName]; ok {
			fileHandle = logFiles[fileName]
		} else {
			// Open the file
			tmp, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("unable to open log file %s: %s", fileName, err.Error()))
			}

			fileHandle = tmp
			logFiles[fileName] = tmp
		}
	}

	// Set log flag
	logFlag := log.Lmsgprefix
	if debug {
		logFlag = logFlag | log.Lshortfile
	}

	// Create and return the logger
	return log.New(logWriter{fileHandle: fileHandle, console: console, debug: debug}, "["+loggerName+"] ", logFlag), nil
}

// Define custom logWriter to control time stamp format
type logWriter struct {
	fileHandle *os.File
	console    bool
	debug      bool
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	event := fmt.Sprint(time.Now().UTC().Format("2006-01-02 15:04:05.000") + " " + string(bytes))
	var n = 0
	var err error = nil

	if writer.console {
		n, err = os.Stdout.WriteString(event)
	}

	if writer.fileHandle != nil {
		n, err = writer.fileHandle.WriteString(event)
	}

	// Return the details of the last write
	return n, err
}
