package posture

import (
	"fmt"

	"github.com/openziti/edge-api/rest_model"
	edge_apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti/edge"
)

// Submitter handles transmission of posture response data to authentication and policy
// enforcement endpoints.
type Submitter interface {
	SendPostureResponse(response rest_model.PostureResponseCreate) error
	SendPostureResponseBulk(responses []rest_model.PostureResponseCreate) error
}

// RouterConnectionProvider supplies active router connections for submitting posture data
// directly to edge routers in high-availability deployments.
type RouterConnectionProvider interface {
	GetRouterConnections() []edge.RouterConn
}

// ApiSessionProvider supplies the current API session, enabling submitters to determine
// the appropriate destination for posture responses based on authentication type.
type ApiSessionProvider interface {
	GetCurrentApiSession() edge_apis.ApiSession
}

var _ Submitter = (*MultiSubmitter)(nil)

// MultiSubmitter routes posture responses to appropriate destinations based on session type
// and router capabilities. Legacy sessions always submit to the controller, while OIDC sessions
// submit to routers that support posture checks and fall back to the controller for older routers.
type MultiSubmitter struct {
	ApiSessionProvider       ApiSessionProvider
	LegacySubmitter          Submitter
	RouterConnectionProvider RouterConnectionProvider
}

// NewMultiSubmitter creates a submitter that intelligently routes posture responses based on
// session authentication method and router capabilities.
func NewMultiSubmitter(apiSessionProvider ApiSessionProvider, legacySubmitter Submitter, routerConnectionProvider RouterConnectionProvider) *MultiSubmitter {
	return &MultiSubmitter{
		ApiSessionProvider:       apiSessionProvider,
		LegacySubmitter:          legacySubmitter,
		RouterConnectionProvider: routerConnectionProvider,
	}
}

func (m *MultiSubmitter) SendPostureResponse(response rest_model.PostureResponseCreate) error {
	if response == nil {
		return nil
	}
	return m.SendPostureResponseBulk([]rest_model.PostureResponseCreate{response})
}

func (m *MultiSubmitter) SendPostureResponseBulk(responses []rest_model.PostureResponseCreate) error {
	if len(responses) == 0 {
		return nil
	}

	apiSession := m.ApiSessionProvider.GetCurrentApiSession()

	//legacy api sessions do not use router posture always goes to the controller
	if apiSession.GetType() == edge_apis.ApiSessionTypeLegacy {
		return m.LegacySubmitter.SendPostureResponseBulk(responses)
	}

	sendToController := false

	routerConns := m.RouterConnectionProvider.GetRouterConnections()
	errors := &MultiDestinationError{
		routerErrors:    map[edge.RouterConn]error{},
		controllerError: nil,
	}

	for _, routerConn := range routerConns {
		if routerConn.GetBoolHeader(edge.SupportsPostureChecksHeader) {
			err := routerConn.SendPosture(responses)
			if err != nil {
				errors.routerErrors[routerConn] = err
			}
		} else {
			sendToController = true
		}
	}

	if sendToController {
		legacyResponse := filterToLegacyPostureResponses(responses)

		errors.controllerError = m.LegacySubmitter.SendPostureResponseBulk(legacyResponse)

	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// filterToLegacyPostureResponses removes responses from the provided list, returning
// only posture response types supported by older controller versions that do not handle
// newer types supported by routers.
func filterToLegacyPostureResponses(responses []rest_model.PostureResponseCreate) []rest_model.PostureResponseCreate {
	legacyResponse := make([]rest_model.PostureResponseCreate, 0, len(responses))

	for _, response := range responses {
		if response.TypeID() != edge.PostureCheckTypeTOTP {
			legacyResponse = append(legacyResponse, response)
		}
	}

	return legacyResponse
}

// MultiDestinationError aggregates errors from posture response submission attempts
// to multiple destinations (controller and routers), providing detailed failure information.
type MultiDestinationError struct {
	routerErrors    map[edge.RouterConn]error
	controllerError error
}

// Error formats all submission failures into a comprehensive error message identifying
// which destinations failed and why.
func (e *MultiDestinationError) Error() string {
	result := ""

	if !e.HasErrors() {
		if e == nil {
			return ""
		}
		panic("unexpected error state, there are no errors, but treated as an error")
	}

	if e.controllerError != nil {
		result = "failed to send posture response to controller: " + e.controllerError.Error()
	}

	if len(e.routerErrors) > 0 {
		if result != "" {
			result += " and "
		}

		routerErrStr := ""

		for routerConn, err := range e.routerErrors {
			if routerErrStr != "" {
				routerErrStr += ", "
			}

			routerErrStr = routerErrStr + fmt.Sprintf("router [%s]: %s", routerConn.GetRouterName(), err.Error())
		}

		result = result + fmt.Sprintf("failed to send posture response to %d routers: %s", len(e.routerErrors), routerErrStr)
	}

	return result
}

// HasErrors returns true if any submission attempts failed, either to the controller
// or to any routers.
func (e *MultiDestinationError) HasErrors() bool {
	return len(e.routerErrors) > 0 || e.controllerError != nil
}
