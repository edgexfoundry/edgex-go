//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	rmFileSuffix string = ".tmp"
)

type fileLog struct {
	filename string
	out      io.WriteCloser
}

func (fl *fileLog) closeSession() {
	if fl.out != nil {
		fl.out.Close()
	}
}

func (fl *fileLog) add(le models.LogEntry) error {
	if fl.out == nil {
		var err error
		//First check to see if the specified directory exists
		//File won't be written without directory.
		path := filepath.Dir(fl.filename)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
		fl.out, err = os.OpenFile(fl.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			//fmt.Println("Error opening log file: ", fl.filename, err)
			fl.out = nil
			return err
		}
	}

	res, err := json.Marshal(le)
	if err != nil {
		return err
	}
	fl.out.Write(res)
	fl.out.Write([]byte("\n"))

	return nil
}

func (fl *fileLog) remove(criteria matchCriteria) (int, error) {
	tmpFilename := fl.filename + rmFileSuffix
	tmpFile, err := os.OpenFile(tmpFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		//fmt.Println("Error creating tmp log file: ", tmpFilename, err)
		return 0, err
	}

	defer os.Remove(tmpFilename)

	f, err := os.Open(fl.filename)
	if err != nil {
		fmt.Println("Error opening log file: ", fl.filename, err)
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
			if !criteria.match(le) {
				tmpFile.Write(line)
				tmpFile.Write([]byte("\n"))
			} else {
				count += 1
			}
		}
	}

	tmpFile.Close()
	err = os.Rename(tmpFilename, fl.filename)
	if err != nil {
		//fmt.Printf("Error renaming %s to %s: %v", tmpFilename, fl.filename, err)
		return 0, err
	}

	// Close old file to open the new one when writing next log
	fl.out.Close()
	fl.out = nil
	return count, nil
}

func (fl *fileLog) find(criteria matchCriteria) ([]models.LogEntry, error) {
	var logs []models.LogEntry
	f, err := os.Open(fl.filename)
	if err != nil {
		//fmt.Println("Error opening log file: ", fl.filename, err)
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var le models.LogEntry

		line := scanner.Bytes()
		err := json.Unmarshal(line, &le)
		if err == nil {
			if criteria.match(le) {
				logs = append(logs, le)

				if criteria.Limit != 0 && len(logs) >= criteria.Limit {
					break
				}
			}
		}
	}
	return logs, err
}

func (fl *fileLog) reset() {
	if fl.out != nil {
		fl.out.Close()
		fl.out = nil
	}
	os.Remove(fl.filename)
}
