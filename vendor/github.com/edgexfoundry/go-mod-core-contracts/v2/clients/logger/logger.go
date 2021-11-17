/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

/*
Package logger provides a client for integration with the support-logging service. The client can also be configured
to write logs to a local file rather than sending them to a service.
*/
package logger

// Logging client for the Go implementation of edgexfoundry

import (
	"fmt"
	stdLog "log"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/go-kit/kit/log"
)

// LoggingClient defines the interface for logging operations.
type LoggingClient interface {
	// SetLogLevel sets minimum severity log level. If a logging method is called with a lower level of severity than
	// what is set, it will result in no output.
	SetLogLevel(logLevel string) errors.EdgeX
	// LogLevel returns the current log level setting
	LogLevel() string
	// Debug logs a message at the DEBUG severity level
	Debug(msg string, args ...interface{})
	// Error logs a message at the ERROR severity level
	Error(msg string, args ...interface{})
	// Info logs a message at the INFO severity level
	Info(msg string, args ...interface{})
	// Trace logs a message at the TRACE severity level
	Trace(msg string, args ...interface{})
	// Warn logs a message at the WARN severity level
	Warn(msg string, args ...interface{})
	// Debugf logs a formatted message at the DEBUG severity level
	Debugf(msg string, args ...interface{})
	// Errorf logs a formatted message at the ERROR severity level
	Errorf(msg string, args ...interface{})
	// Infof logs a formatted message at the INFO severity level
	Infof(msg string, args ...interface{})
	// Tracef logs a formatted message at the TRACE severity level
	Tracef(msg string, args ...interface{})
	// Warnf logs a formatted message at the WARN severity level
	Warnf(msg string, args ...interface{})
}

type edgeXLogger struct {
	owningServiceName string
	logLevel          *string
	rootLogger        log.Logger
	levelLoggers      map[string]log.Logger
}

// NewClient creates an instance of LoggingClient
func NewClient(owningServiceName string, logLevel string) LoggingClient {
	if !isValidLogLevel(logLevel) {
		logLevel = models.InfoLog
	}

	// Set up logging client
	lc := edgeXLogger{
		owningServiceName: owningServiceName,
		logLevel:          &logLevel,
	}

	lc.rootLogger = log.NewLogfmtLogger(os.Stdout)
	lc.rootLogger = log.WithPrefix(
		lc.rootLogger,
		"ts",
		log.DefaultTimestampUTC,
		"app",
		owningServiceName,
		"source",
		log.Caller(5))

	// Set up the loggers
	lc.levelLoggers = map[string]log.Logger{}

	for _, logLevel := range logLevels() {
		lc.levelLoggers[logLevel] = log.WithPrefix(lc.rootLogger, "level", logLevel)
	}

	return lc
}

// LogLevels returns an array of the possible log levels in order from most to least verbose.
func logLevels() []string {
	return []string{
		models.TraceLog,
		models.DebugLog,
		models.InfoLog,
		models.WarnLog,
		models.ErrorLog}
}

func isValidLogLevel(l string) bool {
	for _, name := range logLevels() {
		if name == l {
			return true
		}
	}
	return false
}

func (lc edgeXLogger) log(logLevel string, formatted bool, msg string, args ...interface{}) {
	// Check minimum log level
	for _, name := range logLevels() {
		if name == *lc.logLevel {
			break
		}
		if name == logLevel {
			return
		}
	}

	if args == nil {
		args = []interface{}{"msg", msg}
	} else if formatted {
		args = []interface{}{"msg", fmt.Sprintf(msg, args...)}
	} else {
		if len(args)%2 == 1 {
			// add an empty string to keep k/v pairs correct
			args = append(args, "")
		}
		if len(msg) > 0 {
			args = append(args, "msg", msg)
		}
	}

	err := lc.levelLoggers[logLevel].Log(args...)
	if err != nil {
		stdLog.Fatal(err.Error())
		return
	}

}

func (lc edgeXLogger) SetLogLevel(logLevel string) errors.EdgeX {
	if isValidLogLevel(logLevel) == true {
		*lc.logLevel = logLevel

		return nil
	}

	return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid log level `%s`", logLevel), nil)
}

func (lc edgeXLogger) LogLevel() string {
	if lc.logLevel == nil {
		return ""
	}
	return *lc.logLevel
}

func (lc edgeXLogger) Info(msg string, args ...interface{}) {
	lc.log(models.InfoLog, false, msg, args...)
}

func (lc edgeXLogger) Trace(msg string, args ...interface{}) {
	lc.log(models.TraceLog, false, msg, args...)
}

func (lc edgeXLogger) Debug(msg string, args ...interface{}) {
	lc.log(models.DebugLog, false, msg, args...)
}

func (lc edgeXLogger) Warn(msg string, args ...interface{}) {
	lc.log(models.WarnLog, false, msg, args...)
}

func (lc edgeXLogger) Error(msg string, args ...interface{}) {
	lc.log(models.ErrorLog, false, msg, args...)
}

func (lc edgeXLogger) Infof(msg string, args ...interface{}) {
	lc.log(models.InfoLog, true, msg, args...)
}

func (lc edgeXLogger) Tracef(msg string, args ...interface{}) {
	lc.log(models.TraceLog, true, msg, args...)
}

func (lc edgeXLogger) Debugf(msg string, args ...interface{}) {
	lc.log(models.DebugLog, true, msg, args...)
}

func (lc edgeXLogger) Warnf(msg string, args ...interface{}) {
	lc.log(models.WarnLog, true, msg, args...)
}

func (lc edgeXLogger) Errorf(msg string, args ...interface{}) {
	lc.log(models.ErrorLog, true, msg, args...)
}

// Build the log entry object
func (lc edgeXLogger) buildLogEntry(logLevel string, msg string, args ...interface{}) models.LogEntry {
	res := models.LogEntry{}
	res.Level = logLevel
	res.Message = msg
	res.Args = args
	res.OriginService = lc.owningServiceName

	return res
}
