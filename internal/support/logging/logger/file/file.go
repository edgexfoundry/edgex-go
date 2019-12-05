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
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/support/logging/filter"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Logger struct {
	mutex    sync.Mutex
	filename string
	file     io.WriteCloser
}

func NewLogger(filename string) *Logger {
	return &Logger{
		filename: filename,
	}
}

func (l *Logger) openSession() (err error) {
	if l.file == nil {
		//First check to see if the specified directory exists
		//File won't be written without directory.
		path := filepath.Dir(l.filename)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
		l.file, err = os.OpenFile(l.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			l.file = nil
		}
	}
	return
}

func (l *Logger) closeSession() {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			l.file = nil
		}
	}
}

func (l *Logger) CloseSession() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.closeSession()
}

func (l *Logger) Add(le models.LogEntry) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := l.openSession(); err != nil {
		return err
	}

	res, err := json.Marshal(le)
	if err != nil {
		return err
	}

	l.file.Write(res)
	l.file.Write([]byte("\n"))
	return nil
}

func (l *Logger) Remove(criteria filter.Criteria) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tmpFilename := l.filename + ".tmp"
	tmpFile, err := os.OpenFile(tmpFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}

	l.closeSession()
	f, err := os.Open(l.filename)
	if err != nil {
		fmt.Println("Error opening log file: ", l.filename, err)
		tmpFile.Close()
		os.Remove(tmpFilename)
		return 0, err
	}

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
	f.Close()
	err = os.Rename(tmpFilename, l.filename)
	if err != nil {
		os.Remove(tmpFilename)
		return 0, err
	}

	return count, nil
}

func (l *Logger) Find(criteria filter.Criteria) ([]models.LogEntry, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var logs []models.LogEntry
	f, err := os.Open(l.filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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
