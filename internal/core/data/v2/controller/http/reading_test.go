package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/v2/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingTotalCount(t *testing.T) {
	expectedReadingCount := uint32(656672)
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingTotalCount").Return(expectedReadingCount, nil)

	dic := mocks.NewMockDIC()
	dic.Update(di.ServiceConstructorMap{
		v2DataContainer.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})
	rc := NewReadingController(dic)

	req, err := http.NewRequest(http.MethodGet, v2.ApiReadingCountRoute, http.NoBody)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(rc.ReadingTotalCount)
	handler.ServeHTTP(recorder, req)

	var actualResponse common.CountResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	assert.Equal(t, v2.ApiVersion, actualResponse.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusOK, int(actualResponse.StatusCode), "Response status code not as expected")
	assert.Empty(t, actualResponse.Message, "Message should be empty when it is successful")
	assert.Equal(t, expectedReadingCount, actualResponse.Count, "Event count in the response body is not expected")
}
