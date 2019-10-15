package telemetry

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestNewSystemUsageMemory(t *testing.T) {
	usageUnderTest := NewSystemUsage()
	emptyMemoryUsage := memoryUsage{}

	if usageUnderTest.Memory == emptyMemoryUsage {
		t.Error("Memory usage data was not taken")
	}
}

func TestNewSystemUsageAvg(t *testing.T) {
	var testValue = 13.37
	usageAvg = testValue
	usageUnderTest := NewSystemUsage()

	if usageUnderTest.CpuBusyAvg != testValue {
		t.Error("CpuBusyAvg not correctly copied")
	}
}

func TestCpuUsageAverage(t *testing.T) {
	initialAvg := 13.37
	usageAvg = initialAvg

	// this test is broad because these conditions should be true no matter the OS the test is run on.
	// either it will correctly calculate usage and it will change,
	// or it will run the unimplemented version and change to -1
	cpuUsageAverage()

	if usageAvg == initialAvg {
		t.Fatalf("Expected CPU usageAvg to change, no change. initial: %f, final: %f", initialAvg, usageAvg)
	}
}
