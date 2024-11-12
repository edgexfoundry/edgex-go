//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	pkgUtils "github.com/edgexfoundry/edgex-go/internal/pkg/utils"
)

func sendRESTRequest(lc logger.LoggingClient, action models.RESTAction, jwtSecretProvider interfaces.AuthenticationInjector) (res string, err errors.EdgeX) {
	req, err := getHttpRequestFromRESTAction(action)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "failed to create http request", err)
	}

	if jwtSecretProvider != nil {
		if err2 := jwtSecretProvider.AddAuthenticationData(req); err2 != nil {
			return "", errors.NewCommonEdgeXWrapper(err2)
		}
	}

	client := &http.Client{}
	res, err = pkgUtils.SendRequestAndGetResponse(client, req)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	lc.Debugf("Successfully send the rest request with address %v", action.Address)
	return res, nil
}

func getHttpRequestFromRESTAction(action models.RESTAction) (*http.Request, errors.EdgeX) {
	if !pkgUtils.ValidMethod(action.Method) {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("net/http: invalid method %q", action.Method), nil)
	}

	var body []byte
	if len(action.Payload) > 0 {
		body = action.Payload
	} else {
		body = nil
	}

	req, err := http.NewRequest(action.Method, action.Address, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create new request", err)
	}

	contentType := action.ContentType
	if contentType == "" {
		contentType = common.ContentTypeJSON
	}
	req.Header.Set(common.ContentType, contentType)

	if len(body) > 0 {
		req.Header.Set(common.ContentLength, strconv.Itoa(len(body)))
	}

	return req, nil
}
