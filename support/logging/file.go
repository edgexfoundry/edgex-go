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

	"github.com/edgexfoundry/edgex-go/support/domain"
)

const (
	rmFileSuffix string = ".tmp"
)

type fileLog struct {
	filename string
	out      io.WriteCloser
}

func (fl *fileLog) add(le support_domain.LogEntry) {
	if fl.out == nil {
		var err error
		fl.out, err = os.OpenFile(fl.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			//fmt.Println("Error opening log file: ", fl.filename, err)
			fl.out = nil
			return
		}
	}

	res, err := json.Marshal(le)
	if err != nil {
		return
	}
	fl.out.Write(res)
	fl.out.Write([]byte("\n"))
}

func (fl *fileLog) remove(criteria matchCriteria) int {
	tmpFilename := fl.filename + rmFileSuffix
	tmpFile, err := os.OpenFile(tmpFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		//fmt.Println("Error creating tmp log file: ", tmpFilename, err)
		return 0
	}

	defer os.Remove(tmpFilename)

	f, err := os.Open(fl.filename)
	if err != nil {
		fmt.Println("Error opening log file: ", fl.filename, err)
		tmpFile.Close()
		return 0
	}
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var le support_domain.LogEntry

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
		return 0
	}

	// Close old file to open the new one when writting next log
	fl.out.Close()
	fl.out = nil
	return count
}

func (fl *fileLog) find(criteria matchCriteria) []support_domain.LogEntry {
	var logs []support_domain.LogEntry
	f, err := os.Open(fl.filename)
	if err != nil {
		//fmt.Println("Error opening log file: ", fl.filename, err)
		return nil
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var le support_domain.LogEntry

		line := scanner.Bytes()
		err := json.Unmarshal(line, &le)
		if err == nil {
			if criteria.match(le) {
				logs = append(logs, le)
			}
		}
	}
	return logs
}

func (fl *fileLog) reset() {
	if fl.out != nil {
		fl.out.Close()
		fl.out = nil
	}
	os.Remove(fl.filename)
}
