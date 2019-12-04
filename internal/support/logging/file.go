//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"
)

const (
	rmFileSuffix string = ".tmp"
)

type fileLog struct {
	filename string
	out      io.WriteCloser
}

func (fl *fileLog) CloseSession() {
	if fl.out != nil {
		_ = fl.out.Close()
	}
}

func (fl *fileLog) Add(le models.LogEntry) error {
	if fl.out == nil {
		var err error
		// First check to see if the specified directory exists
		// File won't be written without directory.
		path := filepath.Dir(fl.filename)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			_ = os.MkdirAll(path, 0755)
		}
		fl.out, err = os.OpenFile(fl.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fl.out = nil
			return err
		}
	}

	res, err := json.Marshal(le)
	if err != nil {
		return err
	}
	_, _ = fl.out.Write(res)
	_, _ = fl.out.Write([]byte("\n"))

	return nil
}

func (fl *fileLog) Remove(criteria interfaces.Criteria) (int, error) {
	tmpFilename := fl.filename + rmFileSuffix
	tmpFile, err := os.OpenFile(tmpFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}

	defer os.Remove(tmpFilename)

	f, err := os.Open(fl.filename)
	if err != nil {
		fmt.Println("Error opening log file: ", fl.filename, err)
		_ = tmpFile.Close()
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
				_, _ = tmpFile.Write(line)
				_, _ = tmpFile.Write([]byte("\n"))
			} else {
				count += 1
			}
		}
	}

	_ = tmpFile.Close()
	err = os.Rename(tmpFilename, fl.filename)
	if err != nil {
		return 0, err
	}

	// Close old file to open the new one when writing next log
	_ = fl.out.Close()
	fl.out = nil
	return count, nil
}

func (fl *fileLog) Find(criteria interfaces.Criteria) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	f, err := os.Open(fl.filename)
	if err != nil {
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

				match, ok := criteria.(MatchCriteria)
				if !ok {
					return nil, errors.New("unknown Criteria implementation")
				}

				if match.LimitExceeded(len(logs)) {
					break
				}
			}
		}
	}
	return logs, err
}

