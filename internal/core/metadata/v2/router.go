package v2

import (
	"net/http"

	httpController "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/di"
	v2Constant "github.com/edgexfoundry/go-mod-core-contracts/v2"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Device Profile
	dc := httpController.NewDeviceProfileController(dic)

	r.HandleFunc(v2Constant.ApiDeviceProfileRoute, dc.AddDeviceProfile).Methods(http.MethodPost)
	dp := r.PathPrefix(v2Constant.ApiDeviceProfileRoute).Subrouter()
	dp.HandleFunc(httpController.UploadFile, dc.AddDeviceProfileByYaml).Methods(http.MethodPost)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
