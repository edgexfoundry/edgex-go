// Package logging provides a wrapper around logrus with additional features
//
// Deprecated: This package has been superseded by the standard slog package. Use [slog] instead.
package logging

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

type Entry struct {
	*logrus.Entry
	isOnError bool
	err       error
}

var idKey = "logID"

// SetIDKey key of id in logentry
func SetIDKey(key string) {
	idKey = key
}

// Deprecated: Log creates a new entry with an id
func Log(id string) *Entry {
	entry := (*logrus.Logger)(log).WithField(idKey, id)
	entry.Logger = (*logrus.Logger)(log)
	return &Entry{Entry: entry}
}

// Deprecated: LogWithFields creates a new entry with an id and the given fields
func LogWithFields(id string, fields ...interface{}) *Entry {
	e := Log(id)
	return e.SetFields(fields...)
}

// New instantiates a new entry
func New() *Entry {
	return &Entry{Entry: logrus.NewEntry((*logrus.Logger)(log))}
}

func OnError(err error) *Entry {
	e := New()
	return e.OnError(err)
}

func WithError(err error) *Entry {
	e := New()
	return e.WithError(err)
}

// WithFields creates a new entry without an id and the given fields
func WithFields(fields ...interface{}) *Entry {
	return New().SetFields(fields...)
}

// OnError sets the error. The log will only be printed if err is not nil
func (e *Entry) OnError(err error) *Entry {
	e.err = err
	e.isOnError = true
	return e
}

// SetFields sets the given fields on the entry. It panics if length of fields is odd
func (e *Entry) SetFields(fields ...interface{}) *Entry {
	logFields := toFields(fields...)
	return e.WithFields(logFields)
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	e.Entry = e.Entry.WithField(key, value)
	return e
}

func (e *Entry) WithFields(fields logrus.Fields) *Entry {
	e.Entry = e.Entry.WithFields(fields)
	return e
}

func (e *Entry) WithError(err error) *Entry {
	e.Entry = e.Entry.WithError(err)
	return e
}

func (e *Entry) WithTime(t time.Time) *Entry {
	e.Entry = e.Entry.WithTime(t)
	return e
}

func toFields(fields ...interface{}) logrus.Fields {
	if len(fields)%2 != 0 {
		return logrus.Fields{"oddFields": len(fields)}
	}
	logFields := make(logrus.Fields, len(fields)%2)
	for i := 0; i < len(fields); i = i + 2 {
		key := fields[i].(string)
		logFields[key] = fields[i+1]
	}
	return logFields
}

func Debug(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Debug(args...) })
}

func (e *Entry) Debug(args ...interface{}) {
	e.log(func() { e.Entry.Debug(args...) })
}

func Debugln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Debugln(args...) })
}

func (e *Entry) Debugln(args ...interface{}) {
	e.log(func() { e.Entry.Debugln(args...) })
}

func Debugf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Debugf(format, args...) })
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.log(func() { e.Entry.Debugf(format, args...) })
}

func Info(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Info(args...) })
}

func (e *Entry) Info(args ...interface{}) {
	e.log(func() { e.Entry.Info(args...) })
}

func Infoln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Infoln(args...) })
}

func (e *Entry) Infoln(args ...interface{}) {
	e.log(func() { e.Entry.Infoln(args...) })
}

func Infof(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Infof(format, args...) })
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.log(func() { e.Entry.Infof(format, args...) })
}

func Trace(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Trace(args...) })
}

func (e *Entry) Trace(args ...interface{}) {
	e.log(func() { e.Entry.Trace(args...) })
}

func Traceln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Traceln(args...) })
}

func (e *Entry) Traceln(args ...interface{}) {
	e.log(func() { e.Entry.Traceln(args...) })
}

func Tracef(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Tracef(format, args...) })
}

func (e *Entry) Tracef(format string, args ...interface{}) {
	e.log(func() { e.Entry.Tracef(format, args...) })
}

func Warn(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Warn(args...) })
}

func (e *Entry) Warn(args ...interface{}) {
	e.log(func() { e.Entry.Warn(args...) })
}

func Warnln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Warnln(args...) })
}

func (e *Entry) Warnln(args ...interface{}) {
	e.log(func() { e.Entry.Warnln(args...) })
}

func Warnf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Warnf(format, args...) })
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.log(func() { e.Entry.Warnf(format, args...) })
}

func Warning(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Warning(args...) })
}

func (e *Entry) Warning(args ...interface{}) {
	e.log(func() { e.Entry.Warning(args...) })
}

func Warningln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Warningln(args...) })
}

func (e *Entry) Warningln(args ...interface{}) {
	e.log(func() { e.Entry.Warningln(args...) })
}

func Warningf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Warningf(format, args...) })
}

func (e *Entry) Warningf(format string, args ...interface{}) {
	e.log(func() { e.Entry.Warningf(format, args...) })
}

func Error(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Error(args...) })
}

func (e *Entry) Error(args ...interface{}) {
	e.log(func() { e.Entry.Error(args...) })
}

func Errorln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Errorln(args...) })
}

func (e *Entry) Errorln(args ...interface{}) {
	e.log(func() { e.Entry.Errorln(args...) })
}

func Errorf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Errorf(format, args...) })
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.log(func() { e.Entry.Errorf(format, args...) })
}

func Fatal(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Fatal(args...) })
}

func (e *Entry) Fatal(args ...interface{}) {
	e.log(func() { e.Entry.Fatal(args...) })
}

func Fatalln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Fatalln(args...) })
}

func (e *Entry) Fatalln(args ...interface{}) {
	e.log(func() { e.Entry.Fatalln(args...) })
}

func Fatalf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Fatalf(format, args...) })
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.log(func() { e.Entry.Fatalf(format, args...) })
}

func Panic(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Panic(args...) })
}

func (e *Entry) Panic(args ...interface{}) {
	e.log(func() { e.Entry.Panic(args...) })
}

func Panicln(args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Panicln(args...) })
}

func (e *Entry) Panicln(args ...interface{}) {
	e.log(func() { e.Entry.Panic(args...) })
}

func Panicf(format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Panicf(format, args...) })
}

func (e *Entry) Panicf(format string, args ...interface{}) {
	e.log(func() { e.Entry.Panicf(format, args...) })
}

func (e *Entry) Log(level logrus.Level, args ...interface{}) {
	e.log(func() { e.Entry.Log(level, args...) })
}

func Logf(level logrus.Level, format string, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Logf(level, format, args...) })
}

func (e *Entry) Logf(level logrus.Level, format string, args ...interface{}) {
	e.log(func() { e.Entry.Logf(level, format, args...) })
}

func Logln(level logrus.Level, args ...interface{}) {
	e := New()
	e.log(func() { e.Entry.Logln(level, args...) })
}

func (e *Entry) Logln(level logrus.Level, args ...interface{}) {
	e.log(func() { e.Entry.Logln(level, args...) })
}

func (e *Entry) log(log func()) {
	e = e.checkOnError()
	if e == nil {
		return
	}
	addCaller(e)
	log()
}

func (e *Entry) checkOnError() *Entry {
	if !e.isOnError {
		return e
	}
	if e.err != nil {
		return e.WithError(e.err)
	}
	return nil
}

func addCaller(e *Entry) {
	_, file, no, ok := runtime.Caller(3)
	if ok {
		e.WithField("caller", fmt.Sprintf("%s:%d", file, no))
	}
}
