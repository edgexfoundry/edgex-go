package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

const (
	// URL PATH
	UploadFile = "/uploadfile"
	Name       = "name"
)

type DeviceProfileController struct {
	reader io.DeviceProfileReader
	dic    *di.Container
}

// NewDeviceProfileController creates and initializes an DeviceProfileController
func NewDeviceProfileController(dic *di.Container) *DeviceProfileController {
	return &DeviceProfileController{
		reader: io.NewDeviceProfileRequestReader(),
		dic:    dic,
	}
}

func (dc *DeviceProfileController) AddDeviceProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addDeviceProfileDTOs, err := dc.reader.ReadAddDeviceProfileRequest(r.Body, &ctx)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		http.Error(w, err.Message(), err.Code())
		return
	}
	deviceProfiles := requestDTO.AddDeviceProfileReqToDeviceProfileModels(addDeviceProfileDTOs)

	var addResponses []interface{}
	for i, d := range deviceProfiles {
		var addDeviceProfileResponse interface{}
		// get the requestID from AddDeviceProfileDTO
		reqId := addDeviceProfileDTOs[i].RequestID
		newId, err := application.AddDeviceProfile(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addDeviceProfileResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			addDeviceProfileResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				"Add device profiles successfully",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, addDeviceProfileResponse)
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.Header().Set(clients.CorrelationHeader, correlation.FromContext(ctx))
	w.WriteHeader(http.StatusMultiStatus)
	// Encode and send the resp body as JSON format
	pkg.Encode(addResponses, w, lc)
}

func (dc *DeviceProfileController) AddDeviceProfileByYaml(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	var addDeviceProfileResponse interface{}

	deviceProfileDTO, err := dc.reader.ReadDeviceProfileYaml(r)
	if err != nil {
		addDeviceProfileResponse = commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		w.WriteHeader(err.Code())
		pkg.Encode(addDeviceProfileResponse, w, lc)
		return
	}
	deviceProfile := dtos.ToDeviceProfileModels(deviceProfileDTO)

	newId, err := application.AddDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		addDeviceProfileResponse = commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		w.WriteHeader(err.Code())
		pkg.Encode(addDeviceProfileResponse, w, lc)
		return
	}

	addDeviceProfileResponse = commonDTO.NewBaseWithIdResponse(
		"",
		"Add device profiles successfully",
		http.StatusCreated,
		newId)

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.Header().Set(clients.CorrelationHeader, correlation.FromContext(ctx))
	w.WriteHeader(http.StatusCreated)
	// Encode and send the resp body as JSON format
	pkg.Encode(addDeviceProfileResponse, w, lc)
}
