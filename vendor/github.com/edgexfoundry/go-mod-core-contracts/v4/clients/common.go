// Copyright (C) 2025 IOTech Ltd

package clients

import goErrors "errors"

type ClientBaseUrlFunc func() (string, error)

// GetDefaultClientBaseUrlFunc returns a ClientBaseUrlFunc that always returns the provided baseUrl.
func GetDefaultClientBaseUrlFunc(baseUrl string) ClientBaseUrlFunc {
	return func() (string, error) {
		return baseUrl, nil
	}
}

// GetBaseUrl retrieves the base URL using the provided ClientBaseUrlFunc.
func GetBaseUrl(baseUrlFunc ClientBaseUrlFunc) (string, error) {
	if baseUrlFunc == nil {
		return "", goErrors.New("could not find ClientBaseUrlFunc to get base url")
	}
	return baseUrlFunc()
}
