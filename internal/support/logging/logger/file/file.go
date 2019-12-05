//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/filter"
	"io"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	rmFileSuffix string = ".tmp"
)

type Logger struct {
	filename string
	out      io.WriteCloser
}

func NewLogger(filename string) *Logger {
	return &Logger{
		filename: filename,
	}
}

func (l *Logger) CloseSession() {
	if l.out != nil {
		l.out.Close()
	}
}

func (l *Logger) Add(le models.LogEntry) error {
	if l.out == nil {
		var err error
		//First check to see if the specified directory exists
		//File won't be written without directory.
		path := filepath.Dir(l.filename)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
		l.out, err = os.OpenFile(l.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			//fmt.Println("Error opening log file: ", l.filename, err)
			l.out = nil
			return err
		}
	}

	res, err := json.Marshal(le)
	if err != nil {
		return err
	}
	l.out.Write(res)
	l.out.Write([]byte("\n"))

	return nil
}

func (l *Logger) Remove(criteria filter.Criteria) (int, error) {
	tmpFilename := l.filename + rmFileSuffix
	tmpFile, err := os.OpenFile(tmpFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		//fmt.Println("Error creating tmp log file: ", tmpFilename, err)
		return 0, err
	}

	defer os.Remove(tmpFilename)

	f, err := os.Open(l.filename)
	if err != nil {
		fmt.Println("Error opening log file: ", l.filename, err)
		tmpFile.Close()
		return 0, err
	}
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var le models.LogEntry

		line := scanner.Bytes()
		err := json.Unmarshal(line, &le)
		if err == nil {
			if !criteria.Match(le) {
				tmpFile.Write(line)
				tmpFile.Write([]byte("\n"))
			} else {
				count += 1
			}
		}
	}

	tmpFile.Close()
	err = os.Rename(tmpFilename, l.filename)
	if err != nil {
		//fmt.Printf("Error renaming %s to %s: %v", tmpFilename, l.filename, err)
		return 0, err
	}

	// Close old file to open the new one when writing next log
	l.out.Close()
	l.out = nil
	return count, nil
}

func (l *Logger) Find(criteria filter.Criteria) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	f, err := os.Open(l.filename)
	if err != nil {
		//fmt.Println("Error opening log file: ", l.filename, err)
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var le models.LogEntry

		line := scanner.Bytes()
		err := json.Unmarshal(line, &le)
		if err == nil {
			if criteria.Match(le) {
				logs = append(logs, le)

				if criteria.Limit != 0 && len(logs) >= criteria.Limit {
					break
				}
			}
		}
	}
	return logs, err
}

func (l *Logger) Reset() {
	if l.out != nil {
		l.out.Close()
		l.out = nil
	}
	os.Remove(l.filename)
}
