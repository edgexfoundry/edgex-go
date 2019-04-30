// +build windows

//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package telemetry

import (
	"math"
	"syscall"
	"unsafe"
)

var (
	modkernel32        = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes = modkernel32.NewProc("GetSystemTimes")
)

func PollCpu() (cpuSnapshot CpuUsage) {
	var idle, kernel, user FileTime

	getSystemTimes(&idle, &kernel, &user)
	idleFirst := idle.LowDateTime | (idle.HighDateTime << 32)
	kernelFirst := kernel.LowDateTime | (kernel.HighDateTime << 32)
	userFirst := user.LowDateTime | (user.HighDateTime << 32)

	return CpuUsage{
		Busy:  uint64(userFirst),               // linuxSample.Nice + linuxSample.User,
		Idle:  uint64(idleFirst),               //linuxSample.Idle,
		Total: uint64(kernelFirst + userFirst), // linuxSample.Total,
	}

}

func AvgCpuUsage(init, final CpuUsage) (avg float64) {

	idle := float64(final.Idle - init.Idle)
	total := float64(final.Total - init.Total)

	avg = (total - idle) * 100 / total

	if avg < .000001 || math.IsNaN(avg) {
		return 0.0
	}

	return avg
}

// getSystemTimes makes the system call to windows to get the system data
func getSystemTimes(idleTime, kernelTime, userTime *FileTime) bool {
	ret, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(idleTime)),
		uintptr(unsafe.Pointer(kernelTime)),
		uintptr(unsafe.Pointer(userTime)))

	return ret != 0
}

// FILETIME Struct from: http://msdn.microsoft.com/en-us/library/windows/desktop/ms724284.aspx
type FileTime struct {
	// DwLowDateTime from Windows API
	LowDateTime uint32
	// DwHighDateTime from Windows API
	HighDateTime uint32
}
