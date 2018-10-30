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
 *******************************************************************************/
package logger

// Logging client for the Go implementation of edgexfoundry

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// These constants identify the log levels in order of increasing severity.
const (
	TraceLog = "TRACE"
	DebugLog = "DEBUG"
	InfoLog  = "INFO"
	WarnLog  = "WARN"
	ErrorLog = "ERROR"
)

var LogLevels = []string{
	TraceLog,
	DebugLog,
	InfoLog,
	WarnLog,
	ErrorLog}

type LoggingClient interface {
	SetLogLevel(logLevel string) error
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
	logLevel          *string
	stdOutLogger      *log.Logger
	fileLogger        *log.Logger
}

// Create a new logging client for the owning service
func NewClient(owningServiceName string, isRemote bool, logTarget string, logLevel string) LoggingClient {
	if !IsValidLogLevel(logLevel) {
		logLevel = InfoLog
	}

	// Set up logging client
	lc := EdgeXLogger{
		owningServiceName: owningServiceName,
		remoteEnabled:     isRemote,
		logTarget:         logTarget,
		logLevel:          &logLevel,
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

// IsValidLogLevel checks if is a valid log level
func IsValidLogLevel(l string) bool {
	for _, name := range LogLevels {
		if name == l {
			return true
		}
	}
	return false
}

// Send the log out as a REST request
func (lc EdgeXLogger) log(logLevel string, msg string, labels []string) error {
	// Check minimum log level
	for _, name := range LogLevels {
		if name == *lc.logLevel {
			break
		}
		if name == logLevel {
			return nil
		}
	}

	lc.stdOutLogger.SetPrefix(fmt.Sprintf("%s: ", logLevel))
	lc.stdOutLogger.Println(msg)

	if !lc.remoteEnabled {
		// Save to logging file if path was set
		return lc.saveToLogFile(logLevel, msg)
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

// SetLogLevel sets minimum severity log level
func (lc EdgeXLogger) SetLogLevel(logLevel string) error {
	if IsValidLogLevel(logLevel) == true {
		*lc.logLevel = logLevel

		return nil
	}

	return types.ErrNotFound{}
}

// Log an INFO level message
func (lc EdgeXLogger) Info(msg string, labels ...string) error {
	return lc.log(InfoLog, msg, labels)
}

// Log a TRACE level message
func (lc EdgeXLogger) Trace(msg string, labels ...string) error {
	return lc.log(TraceLog, msg, labels)
}

// Log a DEBUG level message
func (lc EdgeXLogger) Debug(msg string, labels ...string) error {
	return lc.log(DebugLog, msg, labels)
}

// Log a WARN level message
func (lc EdgeXLogger) Warn(msg string, labels ...string) error {
	return lc.log(WarnLog, msg, labels)
}

// Log an ERROR level message
func (lc EdgeXLogger) Error(msg string, labels ...string) error {
	return lc.log(ErrorLog, msg, labels)
}

// Build the log entry object
func (lc EdgeXLogger) buildLogEntry(logLevel string, msg string, labels []string) models.LogEntry {
	res := models.LogEntry{}
	res.Level = logLevel
	res.Message = msg
	res.Labels = labels
	res.OriginService = lc.owningServiceName

	return res
}

// Send the log as an http request
func (lc EdgeXLogger) sendLog(logEntry models.LogEntry) error {
	if lc.logTarget == "" {
		return nil
	}

	go func() {
		_, err := clients.PostJsonRequest(lc.logTarget, logEntry)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return nil
}
