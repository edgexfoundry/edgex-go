package http

import (
	"context"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type DeviceServiceClient struct {
	baseUrl string
}

// NewDeviceServiceClient creates an instance of DeviceServiceClient
func NewDeviceServiceClient(baseUrl string) interfaces.DeviceServiceClient {
	return &DeviceServiceClient{
		baseUrl: baseUrl,
	}
}

func (dsc DeviceServiceClient) Add(ctx context.Context, reqs []requests.AddDeviceServiceRequest) (
	res []dtoCommon.BaseWithIdResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, dsc.baseUrl+common.ApiDeviceServiceRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) Update(ctx context.Context, reqs []requests.UpdateDeviceServiceRequest) (
	res []dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PatchRequest(ctx, &res, dsc.baseUrl+common.ApiDeviceServiceRoute, reqs)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) AllDeviceServices(ctx context.Context, labels []string, offset int, limit int) (
	res responses.MultiDeviceServicesResponse, err errors.EdgeX) {
	requestParams := url.Values{}
	if len(labels) > 0 {
		requestParams.Set(common.Labels, strings.Join(labels, common.CommaSeparator))
	}
	requestParams.Set(common.Offset, strconv.Itoa(offset))
	requestParams.Set(common.Limit, strconv.Itoa(limit))
	err = utils.GetRequest(ctx, &res, dsc.baseUrl, common.ApiAllDeviceServiceRoute, requestParams)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) DeviceServiceByName(ctx context.Context, name string) (
	res responses.DeviceServiceResponse, err errors.EdgeX) {
	path := path.Join(common.ApiDeviceServiceRoute, common.Name, url.QueryEscape(name))
	err = utils.GetRequest(ctx, &res, dsc.baseUrl, path, nil)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}

func (dsc DeviceServiceClient) DeleteByName(ctx context.Context, name string) (
	res dtoCommon.BaseResponse, err errors.EdgeX) {
	path := path.Join(common.ApiDeviceServiceRoute, common.Name, url.QueryEscape(name))
	err = utils.DeleteRequest(ctx, &res, dsc.baseUrl, path)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
