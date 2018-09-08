//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package memory

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db/test"
)

func TestMemoryDB(t *testing.T) {
	memory := &MemDB{}
	test.TestDataDB(t, memory)
	test.TestMetadataDB(t, memory)
	test.TestExportDB(t, memory)
}

func BenchmarkMemoryDB(b *testing.B) {
	memory := &MemDB{}

	test.BenchmarkDB(b, memory)
}
