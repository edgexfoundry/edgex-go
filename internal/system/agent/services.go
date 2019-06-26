/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

func InvokeMetrics(services []string, ctx context.Context) (MetricsRespMap, error) {

	m := MetricsRespMap{}
	m.Metrics = map[string]interface{}{}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Info("invoke metrics")

		if Configuration.MetricsMechanism == "direct-service" {
			es := ExecuteService{}
			out, err := es.Metrics(ctx, service)
			if err != nil {
				LoggingClient.Error("error fetching metrics")
				return m, err
			} else {
				m.Metrics[service] = ProcessResponse(string(out))
			}
		} else if Configuration.MetricsMechanism == "executor" {
			var result []byte
			ea := ExecuteApp{}
			out, err := ea.Metrics(ctx, service)
			if err.Error() != "exit status 1" {
				LoggingClient.Error("error fetching metrics")
				return m, err
			} else {
				result, _ = processOutput(out, service)
				m.Metrics[service] = ProcessResponse(string(result))
			}
		} else if Configuration.MetricsMechanism == "custom" {
			err := fmt.Errorf("the requested custom executor (e.g. snap) has not been integrated")
			LoggingClient.Error(err.Error())
			m.Metrics[service] = fmt.Sprintf(err.Error())
		} else {
			err := fmt.Errorf("the requested metrics mechanism is not supported")
			LoggingClient.Error(err.Error())
		}
	}
	return m, nil
}

func processOutput(bytes []byte, service string) ([]byte, error) {

	var relevantLines []string
	s := string(bytes)
	var fields [][]string
	var populated = map[string]Stats{}

	lines, err := stringToLines(s)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}
	relevantLines = findRelevantLines(lines)
	fields = tabulateResult(relevantLines)
	populated = populateMap(fields)
	relevantStats := selectRelevantStats(populated, service)
	response, err := assembleStats(relevantStats)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}
	return response, nil
}

func assembleStats(stats map[string]Stats) ([]byte, error) {

	const (
		CpuPercent       = "CPU %"
		MemUsageAndLimit = "MEM USAGE / LIMIT"
		MemPercent       = "MEM %"
		Net_IO           = "NET I/O"
		Block_IO         = "BLOCK I/O"
		Pids             = "PIDS"
	)

	data := make(map[string]string)

	for _, value := range stats {
		data[CpuPercent] = value.CpuPercent
		data[MemUsageAndLimit] = value.MemUsageAndLimit
		data[MemPercent] = value.MemPercent
		data[Net_IO] = value.Net_IO
		data[Block_IO] = value.Block_IO
		data[Pids] = value.Pids
	}

	assembled, err := json.Marshal(data)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}
	return assembled, err
}

func selectRelevantStats(stats map[string]Stats, service string) map[string]Stats {

	var relevantStats = map[string]Stats{}

	for key, value := range stats {
		if key == service {
			relevantStats[key] = value
			return relevantStats
		}
	}
	return relevantStats
}

type Stats struct {
	CpuPercent       string
	MemUsageAndLimit string
	MemPercent       string
	Net_IO           string
	Block_IO         string
	Pids             string
}

func populateMap(fields [][]string) map[string]Stats {

	data := make(map[string]Stats)

	// Populate the map ("data") with elements of the console output that are
	// retrieved after running the "docker stats" command.
	size := len(fields)
	for i := 0; i < size; i++ {
		data[fields[i][1]] = Stats{
			fields[i][2],
			fields[i][3] + fields[i][4] + fields[i][5],
			fields[i][6],
			fields[i][7] + fields[i][8] + fields[i][9],
			fields[i][10] + fields[i][11] + fields[i][12],
			fields[i][13],
		}
	}
	return data
}

func tabulateResult(relevantLines []string) [][]string {

	var fields [][]string
	for _, line := range relevantLines {
		row := strings.Fields(line)
		fields = append(fields, row)
	}

	return fields
}

func findRelevantLines(allLines []string) []string {

	var relevantLines []string
	var truncated []string
	var proceed bool
	proceed = false
	count := 0

	for _, line := range allLines {
		if proceed && count < 2 {
			relevantLines = append(relevantLines, line)
		} else if count == 2 {
			break
		}

		if strings.Contains(line, "CONTAINER ID") {
			proceed = true
			count = count + 1
		} else if count > 1 || strings.Contains(line, "cmd.Run() failed with signal") {
			proceed = false
		}
	}
	// The following brief logic is to handle an artifact of the way the console output is retrieved
	// after running the command "docker stats".
	truncated = relevantLines[:len(relevantLines)-1]

	return truncated
}

func stringToLines(s string) (lines []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()

	return lines, err
}

func InvokeOperation(action string, services []string) error {

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Info("invoking operation")

		if !IsKnownServiceKey(service) {
			LoggingClient.Warn(fmt.Sprintf("unknown service %s during invocation", service))
		}

		switch action {

		case START:
			if starter, ok := executorClient.(interfaces.ServiceStarter); ok {
				err := starter.Start(service)
				if err != nil {
					LoggingClient.Error("error starting service")
					return err
				}
			} else {
				err := fmt.Errorf("operation not supported with specified executor")
				LoggingClient.Error(err.Error())
				return err
			}

		case STOP:
			if stopper, ok := executorClient.(interfaces.ServiceStopper); ok {
				err := stopper.Stop(service)
				if err != nil {
					LoggingClient.Error("error stopping service")
					return err
				}
			} else {
				err := fmt.Errorf("operation not supported with specified executor")
				LoggingClient.Error(err.Error())
				return err
			}

		case RESTART:
			if restarter, ok := executorClient.(interfaces.ServiceRestarter); ok {
				err := restarter.Restart(service)
				if err != nil {
					LoggingClient.Error("error restarting service")
					return err
				}
			} else {
				err := fmt.Errorf("operation not supported with specified executor")
				LoggingClient.Error(err.Error())
				return err
			}
		}
	}
	return nil
}

func getConfig(services []string, ctx context.Context) (ConfigRespMap, error) {

	c := ConfigRespMap{}
	c.Configuration = map[string]interface{}{}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {

		// Check whether SMA does _not_ know of ServiceKey ("service") as being one for one of its ready-made list of clients.
		if !IsKnownServiceKey(service) {
			LoggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", service))

			// Service unknown to SMA, so ask the Registry whether `service` is available.
			err := registryClient.IsServiceAvailable(service)
			if err != nil {
				c.Configuration[service] = fmt.Sprintf(err.Error())
				LoggingClient.Error(err.Error())
			} else {
				LoggingClient.Info(fmt.Sprintf("Registry responded with %s service available", service))

				// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `service`
				e, err := registryClient.GetServiceEndpoint(service)
				if err != nil {
					c.Configuration[service] = fmt.Sprintf("on attempting to get ServiceEndpoint for service %s, got error: %v", service, err.Error())
					LoggingClient.Error(err.Error())
				} else {
					// Preparing to add the specified key to the map where the value will be the respective GeneralClient
					clientInfo := config.ClientInfo{}
					clientInfo.Protocol = Configuration.Service.Protocol
					clientInfo.Host = e.Host
					clientInfo.Port = e.Port

					// This code will evolve to take into account a manifest-like functionality in future. So
					// rather than assume that the runtime bool flag useRegistry has been initialized to true,
					// given that the flow has reached this point, having already called functions on the Registry,
					// such as registryClient.IsServiceAvailable(service), we test for its truthiness. I expect
					// this code to be refactored as we evolve toward a manifest-like functionality in future.
					usingRegistry := false
					if registryClient != nil {
						usingRegistry = true
					}

					Configuration.Clients[e.ServiceId] = clientInfo
					params := types.EndpointParams{
						ServiceKey:  e.ServiceId,
						Path:        "/",
						UseRegistry: usingRegistry,
						Url:         Configuration.Clients[e.ServiceId].Url() + clients.ApiConfigRoute,
						Interval:    internal.ClientMonitorDefault,
					}
					// TODO: Related to future work: With the current deployment strategy, GeneralClient's
					//  func init(params types.EndpointParams) {...} returns blank "url"...
					// Add the service key to the map where the value is the respective GeneralClient
					generalClients[e.ServiceId] = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &registryClient})

					responseJSON, err := generalClients[e.ServiceId].FetchConfiguration(ctx)
					if err != nil {
						c.Configuration[service] = fmt.Sprintf(err.Error())
						LoggingClient.Error(err.Error())
					} else {
						c.Configuration[service] = ProcessResponse(responseJSON)
					}
					return c, nil
				}
			}
		} else {
			// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `service`
			// Simply use one of the ready-made list of clients.
			LoggingClient.Info(fmt.Sprintf("service %s is known to SMA as being in the ready-made list of clients", service))

			responseJSON, err := generalClients[service].FetchConfiguration(ctx)
			if err != nil {
				c.Configuration[service] = fmt.Sprintf(err.Error())
				LoggingClient.Error(err.Error())
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
		}
	}
	return c, nil
}

type ExecuteService struct {
}

func (ec *ExecuteService) Metrics(ctx context.Context, service string) ([]byte, error) {

	out := []byte("")
	m := MetricsRespMap{}
	m.Metrics = map[string]interface{}{}

	// Check whether SMA does _not_ know of ServiceKey ("service") as being one for one of its ready-made list of clients.
	if !IsKnownServiceKey(service) {
		LoggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", service))

		// Service unknown to SMA, so ask the Registry whether `service` is available.
		err := registryClient.IsServiceAvailable(service)
		if err != nil {
			out = []byte(fmt.Sprintf(err.Error()))
			LoggingClient.Debug(fmt.Sprintf(string(out)))
		} else {
			LoggingClient.Info(fmt.Sprintf("Registry responded with %s service available", service))

			// Since service is unknown to SMA, ask the Registry for a ServiceEndpoint associated with `service`
			e, err := registryClient.GetServiceEndpoint(service)
			if err != nil {
				out = []byte(fmt.Sprintf("on attempting to get ServiceEndpoint for service %s, got error: %v", service, err.Error()))
				LoggingClient.Error(fmt.Sprintf(service, err.Error()))
			} else {
				// Preparing to add the specified key to the map where the value will be the respective GeneralClient
				clientInfo := config.ClientInfo{}
				clientInfo.Protocol = Configuration.Service.Protocol
				clientInfo.Host = e.Host
				clientInfo.Port = e.Port

				// This code will evolve to take into account a manifest-like functionality in future. So
				// rather than assume that the runtime bool flag useRegistry has been initialized to true,
				// given that the flow has reached this point, having already called functions on the Registry,
				// such as registryClient.IsServiceAvailable(service), we test for its truthiness. I expect
				// this code to be refactored as we evolve toward a manifest-like functionality in future.
				usingRegistry := false
				if registryClient != nil {
					usingRegistry = true
				}

				Configuration.Clients[e.ServiceId] = clientInfo
				params := types.EndpointParams{
					ServiceKey:  e.ServiceId,
					Path:        "/",
					UseRegistry: usingRegistry,
					Url:         Configuration.Clients[e.ServiceId].Url() + clients.ApiMetricsRoute,
					Interval:    internal.ClientMonitorDefault,
				}
				// TODO: Related to future work: With the current deployment strategy, GeneralClient's
				//  func init(params types.EndpointParams) {...} returns blank "url"...
				// Add the service key to the map where the value is the respective GeneralClient
				generalClients[e.ServiceId] = general.NewGeneralClient(params, startup.Endpoint{RegistryClient: &registryClient})
			}
		}
	} else {
		// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `service`
		// Simply use one of the ready-made list of clients.
		LoggingClient.Info(fmt.Sprintf("service %s is known to SMA as being in the ready-made list of clients", service))

		responseJSON, err := generalClients[service].FetchMetrics(ctx)
		if err != nil {
			out = []byte(fmt.Sprintf(err.Error()))
			LoggingClient.Error(err.Error())
		} else {
			out = []byte(responseJSON)
		}
	}
	return out, nil
}

func getHealth(services []string) (map[string]interface{}, error) {

	health := make(map[string]interface{})

	for _, service := range services {

		if !IsKnownServiceKey(service) {
			LoggingClient.Warn(fmt.Sprintf("unknown service %s found while getting health", service))
		}

		err := registryClient.IsServiceAvailable(service)
		// the registry service returns nil for a healthy service
		if err != nil {
			health[service] = err.Error()
		} else {
			health[service] = true
		}
	}

	return health, nil
}

func IsKnownServiceKey(serviceKey string) bool {
	// create a map because this is the easiest/cleanest way to determine whether something exists in a set
	var services = map[string]struct{}{
		clients.SupportNotificationsServiceKey: {},
		clients.CoreCommandServiceKey:          {},
		clients.CoreDataServiceKey:             {},
		clients.CoreMetaDataServiceKey:         {},
		clients.ExportClientServiceKey:         {},
		clients.ExportDistroServiceKey:         {},
		clients.SupportLoggingServiceKey:       {},
		clients.SupportSchedulerServiceKey:     {},
		clients.ConfigSeedServiceKey:           {},
	}

	_, exists := services[serviceKey]

	return exists
}
