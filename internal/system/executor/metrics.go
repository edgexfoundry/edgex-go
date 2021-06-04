/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package executor

import (
	"encoding/json"
	"fmt"
	"strings"
)

const separator = ";"

type metricsResult struct {
	CpuUsedPercent float64         `json:"cpuUsedPercent"`
	MemoryUsed     int64           `json:"memoryUsed"`
	Raw            json.RawMessage `json:"raw"`
}

// metricsExecutorCommands returns the Docker command to be executed to gather metrics from the Docker CLI.
func metricsExecutorCommands(serviceName string) []string {
	return []string{
		"stats",
		serviceName,
		"--no-stream",
		"--format",
		"{{ .CPUPerc }}" + separator +
			"{{ .MemUsage }}" + separator +
			"{\"cpu_perc\":\"{{ .CPUPerc }}\",\"mem_usage\":\"{{ .MemUsage }}\",\"mem_perc\":\"{{ .MemPerc }}\",\"net_io\":\"{{ .NetIO }}\",\"block_io\":\"{{ .BlockIO }}\",\"pids\":\"{{ .PIDs }}\"}",
	}
}

// dockerCpuToFloat converts the CPU value returned from the Docker CLI to a float64 value.
func dockerCpuToFloat(value string) float64 {
	var cpu float64
	n, err := fmt.Sscanf(value, "%f", &cpu)
	if err != nil || n != 1 {
		return -1.0
	}
	return cpu
}

// dockerMemoryToInt converts the memory used value returned from the Docker CLI to an int64 value.
func dockerMemoryToInt(value string) int64 {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	var memory float64
	var scale string
	n, err := fmt.Sscanf(value, "%f%s", &memory, &scale)
	if err != nil || n != 2 {
		return -1
	}
	switch scale {
	case "KiB":
		memory *= kb
	case "MiB":
		memory *= mb
	case "GiB":
		memory *= gb
	}
	return int64(memory + 0.5)
}

// resultToFields converts and returns the values returned by the Docker CLI into normalized cpuUsedPercent and
// memoryUsed fields provided by every executor (along with the raw Docker CLI result).
func resultToFields(result string) (cpuUsedPercent float64, memoryUsed int64, raw []byte) {
	resultFields := strings.Split(strings.TrimRight(result, "\n"), separator)
	cpuUsedPercent = dockerCpuToFloat(resultFields[0])
	memoryUsed = dockerMemoryToInt(strings.Split(resultFields[1], " ")[0])
	raw = []byte(resultFields[2])
	return
}

// gatherMetrics delegates metrics gathering to the executor and converts the result to a MetricsSuccessResult.
func gatherMetrics(serviceName string, executor CommandExecutor) (res metricsResult, err error) {
	output, err := executor(metricsExecutorCommands(serviceName)...)
	if err != nil {
		return res, fmt.Errorf("%s: %s", err.Error(), output)
	}

	cpuUsedPercent, memoryUsed, raw := resultToFields(string(output))
	return metricsResult{
		CpuUsedPercent: cpuUsedPercent,
		MemoryUsed:     memoryUsed,
		Raw:            raw,
	}, nil
}
