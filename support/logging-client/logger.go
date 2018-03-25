/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @microservice: support-logging-client-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package logger

// Logging client for the Go implementation of edgexfoundry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"path/filepath"
	"strings"

	"github.com/edgexfoundry/edgex-go/support/domain"
)

type LoggingClient interface {
	Debug(msg string, labels ...string) error
	Error(msg string, labels ...string) error
	Info(msg string, labels ...string) error
	Trace(msg string, labels ...string) error
	Warn(msg string, labels ...string) error
}

type EdgeXLogger struct {
	owningServiceName string
	remoteEnabled     bool
	logTarget         string
	stdOutLogger      *log.Logger
	fileLogger        *log.Logger
}

// Create a new logging client for the owning service
func NewClient(owningServiceName string, isRemote bool, logTarget string) LoggingClient {
	// Set up logging client
	lc := EdgeXLogger{
		owningServiceName: owningServiceName,
		remoteEnabled:     isRemote,
		logTarget:         logTarget,
	}

	//If local logging, verify directory exists
	if !lc.remoteEnabled {
		verifyLogDirectory(lc.logTarget)
	}

	// Set up the loggers
	lc.stdOutLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	lc.fileLogger = &log.Logger{}
	lc.fileLogger.SetFlags(log.Ldate | log.Ltime)

	return lc
}

// Send the log out as a REST request
func (lc EdgeXLogger) log(logLevel string, msg string, labels []string) error {
	if !lc.remoteEnabled {
		// Save to logging file if path was set
		return lc.saveToLogFile(string(logLevel), msg)
	}

	// Send to logging service
	logEntry := lc.buildLogEntry(logLevel, msg, labels)
	return lc.sendLog(logEntry)
}

func (lc EdgeXLogger) saveToLogFile(prefix string, message string) error {
	if lc.logTarget == "" {
		return nil
	}
	file, err := os.OpenFile(lc.logTarget, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		fmt.Println("Error opening log file: " + err.Error())
	}
	lc.fileLogger.SetOutput(file)
	lc.fileLogger.SetPrefix(prefix + ": ")
	lc.fileLogger.Println(message)
	return nil
}

func verifyLogDirectory(path string) {
	prefix, _ := filepath.Split(path)
	//If a path to the log file was specified and it does not exist, create it.
	dir := strings.TrimRight(prefix, "/")
	if len(dir) > 0 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Println("Creating directory: " + dir)
			os.MkdirAll(dir, 0766)
		}
	}
}

// Log an INFO level message
func (lc EdgeXLogger) Info(msg string, labels ...string) error {
	lc.stdOutLogger.SetPrefix("INFO: ")
	lc.stdOutLogger.Println(msg)
	return lc.log(support_domain.INFO, msg, labels)
}

// Log a TRACE level message
func (lc EdgeXLogger) Trace(msg string, labels ...string) error {
	lc.stdOutLogger.SetPrefix("TRACE: ")
	lc.stdOutLogger.Println(msg)
	return lc.log(support_domain.TRACE, msg, labels)
}

// Log a DEBUG level message
func (lc EdgeXLogger) Debug(msg string, labels ...string) error {
	lc.stdOutLogger.SetPrefix("DEBUG: ")
	lc.stdOutLogger.Println(msg)
	return lc.log(support_domain.DEBUG, msg, labels)
}

// Log a WARN level message
func (lc EdgeXLogger) Warn(msg string, labels ...string) error {
	lc.stdOutLogger.SetPrefix("WARN: ")
	lc.stdOutLogger.Println(msg)
	return lc.log(support_domain.WARN, msg, labels)
}

// Log an ERROR level message
func (lc EdgeXLogger) Error(msg string, labels ...string) error {
	lc.stdOutLogger.SetPrefix("ERROR: ")
	lc.stdOutLogger.Println(msg)
	return lc.log(support_domain.ERROR, msg, labels)
}

// Build the log entry object
func (lc EdgeXLogger) buildLogEntry(logLevel string, msg string, labels []string) support_domain.LogEntry {
	res := support_domain.LogEntry{}
	res.Level = logLevel
	res.Message = msg
	res.Labels = labels
	res.OriginService = lc.owningServiceName

	return res
}

// Send the log as an http request
func (lc EdgeXLogger) sendLog(logEntry support_domain.LogEntry) error {
	if lc.logTarget == "" {
		return nil
	}

	reqBody, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	req, err := http.NewRequest(http.MethodPost, lc.logTarget, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	// Asynchronous call
	go lc.makeRequest(client, req)

	return nil
}

// Function to call in a goroutine
func (lc EdgeXLogger) makeRequest(client *http.Client, request *http.Request) {
	resp, err := client.Do(request)
	if err == nil {
		defer resp.Body.Close()
		resp.Close = true
	} else {
		fmt.Println(err.Error())
	}
}
