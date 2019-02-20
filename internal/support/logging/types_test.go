//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
)

func logAllLevels(pl privLogger, msg string) {
	pl.Debug(msg)
	pl.Error(msg)
	pl.Info(msg)
	pl.Trace(msg)
	pl.Warn(msg)
}

func TestPrivLogger(t *testing.T) {
	pl := newPrivateLogger()
	pl.SetLogLevel(logger.TraceLog)

	// Does not break with nil persistence
	dbClient = nil
	logAllLevels(pl, "dummy")

	dp := dummyPersist{}
	dbClient = &dp
	logAllLevels(pl, "dummy")
	if dp.added != 5 {
		t.Errorf("Should be added 5 logs, not %d", dp.added)
	}
}
