package agent

import (
	requests "github.com/edgexfoundry/go-mod-core-contracts/requests/configuration"
	responses "github.com/edgexfoundry/go-mod-core-contracts/responses/configuration"
)

type stubCallSet struct {
	expectedArgsSet []string                    // expected arg value for specific executor call
	outString       responses.SetConfigResponse // return value for specific executor call
}

type expectedArgsSet struct {
	service string
	sc      requests.SetConfigRequest
}

type StubSet struct {
	Called         int               // number of times stub is called
	capturedArgs   []expectedArgsSet // captures arg values for each stub call
	perCallResults stubCallSet       // expected arg value and return values for each stub call
}

func NewStubSet(results stubCallSet) StubSet {
	return StubSet{
		perCallResults: results,
	}
}

// This is a stub implementation of the SetExecutor interface.
func (m *StubSet) SetExecutor(service string,
	sc requests.SetConfigRequest) responses.SetConfigResponse {
	m.Called++
	m.capturedArgs = append(m.capturedArgs, expectedArgsSet{service, sc})
	return m.perCallResults.outString
}
