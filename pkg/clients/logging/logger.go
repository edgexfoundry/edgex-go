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
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/kit/log"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
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
	Debug(msg string, args ...interface{}) error
	Error(msg string, args ...interface{}) error
	Info(msg string, args ...interface{}) error
	Trace(msg string, args ...interface{}) error
	Warn(msg string, args ...interface{}) error
}

type EdgeXLogger struct {
	owningServiceName string
	remoteEnabled     bool
	logTarget         string
	logLevel          *string
	rootLogger        log.Logger
	levelLoggers      map[string]log.Logger
}

type fileWriter struct {
	fileName string
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

	// Set up the loggers
	lc.levelLoggers = map[string]log.Logger{}

	var err error

	//If local logging, verify directory exists
	if !lc.remoteEnabled {
		verifyLogDirectory(lc.logTarget)

		w, err := newFileWriter(lc.logTarget)
		// only use the FileWriter if there are no errors
		if err != nil {
			stdlog.Fatal(err.Error())
		}
		lc.rootLogger = log.NewLogfmtLogger(io.MultiWriter(os.Stdout, log.NewSyncWriter(w)))
	} else {
		lc.rootLogger = log.NewLogfmtLogger(os.Stdout)
	}

	lc.rootLogger = log.WithPrefix(lc.rootLogger, "ts", log.DefaultTimestampUTC,
		"source", log.Caller(5))
	if err != nil {
		lc.Error(err.Error(), "initializationStatus", "error")
	}
	return lc
}

func (f *fileWriter) Write(p []byte) (n int, err error) {
	file, err := os.OpenFile(f.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	_, err = file.WriteString(string(p))
	return len(p), err
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

func newFileWriter(logTarget string) (io.Writer, error) {
	if logTarget == "" {
		return nil, errors.New("logTarget cannot be blank")
	}
	fileWriter := fileWriter{fileName: logTarget}

	return &fileWriter, nil
}

func (lc EdgeXLogger) log(logLevel string, msg string, args ...interface{}) error {
	// Check minimum log level
	for _, name := range LogLevels {
		if name == *lc.logLevel {
			break
		}
		if name == logLevel {
			return nil
		}
	}

	if args == nil {
		args = []interface{}{"msg", msg}
	} else {
		if len(args)%2 == 1 {
			// add an empty string to keep k/v pairs correct
			args = append(args, "")
		}
		// Practical usage thus far has been to call this type like so Logger.Info("message")
		// I'm attempting to preserve that behavior below without requiring the client types
		// to provide the "msg" key.
		args = append(args, "msg", msg)
	}

	if lc.levelLoggers[logLevel] == nil {
		lc.levelLoggers[logLevel] = log.WithPrefix(lc.rootLogger, "level", logLevel)
	}

	err := lc.levelLoggers[logLevel].Log(args...)
	if err != nil {
		return err
	}

	if lc.remoteEnabled {
		// Send to logging service
		logEntry := lc.buildLogEntry(logLevel, msg, args...)
		lc.sendLog(logEntry)
	}
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
func (lc EdgeXLogger) Info(msg string, args ...interface{}) error {
	return lc.log(InfoLog, msg, args...)
}

// Log a TRACE level message
func (lc EdgeXLogger) Trace(msg string, args ...interface{}) error {
	return lc.log(TraceLog, msg, args...)
}

// Log a DEBUG level message
func (lc EdgeXLogger) Debug(msg string, args ...interface{}) error {
	return lc.log(DebugLog, msg, args...)
}

// Log a WARN level message
func (lc EdgeXLogger) Warn(msg string, args ...interface{}) error {
	return lc.log(WarnLog, msg, args...)
}

// Log an ERROR level message
func (lc EdgeXLogger) Error(msg string, args ...interface{}) error {
	return lc.log(ErrorLog, msg, args...)
}

// Build the log entry object
func (lc EdgeXLogger) buildLogEntry(logLevel string, msg string, args ...interface{}) models.LogEntry {
	res := models.LogEntry{}
	res.Level = logLevel
	res.Message = msg
	res.Args = args
	res.OriginService = lc.owningServiceName

	return res
}

// Send the log as an http request
func (lc EdgeXLogger) sendLog(logEntry models.LogEntry) {
	if lc.logTarget == "" {
		return
	}

	go func() {
		_, err := clients.PostJsonRequest(lc.logTarget, logEntry)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
}
