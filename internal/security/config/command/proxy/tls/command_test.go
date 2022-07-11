//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package tls

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyCertCase = iota
	oneCertCase
	twoOrMoreCertCase
)

// TestTLSBagArguments tests command line errors
func TestTLSBagArguments(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	badArgTestcases := [][]string{
		{},                       // missing arg --in
		{"-badarg"},              // invalid arg
		{"--incert", "somefile"}, // missing --inkey
		{"--inkey", "keyfile"},   // missing --incert
	}

	for _, args := range badArgTestcases {
		// Act
		command, err := NewCommand(lc, config, args)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, command)
	}
}

// TestTLSErrorFileNotFound tests the tls error regarding file not found issues
func TestTLSErrorFileNotFound(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}
	fileNotFoundTestcases := [][]string{
		{"--incert", "missingcertificate", "--inkey", "missingprivatekey", "--admin_api_jwt", "random"},       // both files missing
		{"--incert", "testdata/testCert.pem", "--inkey", "missingprivatekey", "--admin_api_jwt", "random"},    // key file missing
		{"--incert", "missingcertificate", "--inkey", "testdata/testCert.prkey", "--admin_api_jwt", "random"}, // cert file missing
	}

	for _, args := range fileNotFoundTestcases {
		// Act
		command, err := NewCommand(lc, config, args)
		require.NoError(t, err)
		code, err := command.Execute()

		// Assert
		require.Error(t, err)
		require.Equal(t, interfaces.StatusCodeExitWithError, code)
	}
}

func TestTLSAddNewCertificate(t *testing.T) {
	// Arrange
	lc := logger.MockLogger{}
	config := &config.ConfigurationStruct{}

	tests := []struct {
		name         string
		certCase     int
		listCertOk   bool
		deleteCertOk bool
		postCertOk   bool
		expectedErr  bool
	}{
		{"Good: Add new kong cert when cert in server is empty", emptyCertCase, true, true, true, false},
		{"Good: Add new kong cert when one cert is in server", oneCertCase, true, true, true, false},
		{"Good: Add new kong cert when two or more cert is in server", twoOrMoreCertCase, true, true, true, false},
		{"Ok: Add new kong cert when delete certificate API failed but list cert being empty", emptyCertCase, true, false, true, false},
		{"Bad: Add new kong cert when list certificate API failed", oneCertCase, false, true, true, true},
		{"Bad: Add new kong cert when delete certificate API failed", oneCertCase, true, false, true, true},
		{"Bad: Add new kong cert when post certificate API failed", oneCertCase, true, true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := getTlsCertificateTestServer(tt.certCase, tt.listCertOk, tt.deleteCertOk, tt.postCertOk, t)
			defer ts.Close()

			tsURL, err := url.Parse(ts.URL)
			require.NoError(t, err)

			config.KongURL.Server = tsURL.Hostname()
			config.KongURL.ApplicationPortSSL, _ = strconv.Atoi(tsURL.Port())

			args := []string{
				"--incert", "testdata/testCert.pem",
				"--inkey", "testdata/testCert.prkey",
				"--admin_api_jwt", "random",
			}

			// Act
			command, err := NewCommand(lc, config, args)
			require.NoError(t, err)
			code, err := command.Execute()

			// Assert
			if tt.expectedErr {
				require.Error(t, err)
				require.Equal(t, interfaces.StatusCodeExitWithError, code)
			} else {
				require.NoError(t, err)
				require.Equal(t, interfaces.StatusCodeExitNormal, code)
			}
		})
	}
}

func TestGetServerNameIndicators(t *testing.T) {
	builtinSnis := []string{"localhost", "kong"}
	tests := []struct {
		name         string
		snisInputStr string
		expectedSnis []string
	}{
		{"Empty SNIS input", "", builtinSnis},
		{"One SNIS input", "test1.com", append(builtinSnis, "test1.com")},
		{"Two SNIS input", "test1.com,test2.com", append(builtinSnis, []string{"test1.com", "test2.com"}...)},
		{"Two or more SNIS with spaces", "test1.com, test2.com, test3.com",
			append(builtinSnis, []string{"test1.com", "test2.com", "test3.com"}...)},
		{"Mixed with empty entries", ", test1.com, ", append(builtinSnis, "test1.com")},
		{"Equivalent empty entries", ",,", builtinSnis},
		{"Duplicate with internal entries", "kong, localhost", builtinSnis},
		{"Mixed with some duplicating internal entries", "kong, test1.com, test2.com,localhost,kong",
			append(builtinSnis, []string{"test1.com", "test2.com"}...)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedSnis, getServerNameIndicators(tt.snisInputStr))
		})
	}
}

func getTlsCertificateTestServer(listCertCase int, listCertOk bool, deleteCertOk bool, postCertOk bool,
	t *testing.T) *httptest.Server {
	builtinSnis := []string{"localhost", "kong"}
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.EscapedPath()
		switch r.Method {
		case http.MethodGet:
			if urlPath == "/admin/snis" {
				if listCertOk {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
				var jsonResponse map[string]interface{}
				switch listCertCase {
				default:
				case emptyCertCase:
					jsonResponse = map[string]interface{}{
						"data": []string{},
						"next": nil,
					}
				case oneCertCase:
					jsonResponse = map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id":        "fake-snis-id-01",
								"createdAt": 1425366534,
								"name":      builtinSnis[0],
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-01",
								},
							},
							{
								"id":        "fake-snis-id-02",
								"createdAt": 1425366534,
								"name":      builtinSnis[1],
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-01",
								},
							},
						},
						"next": "https://localhost:8000/admin/certificates?offset=xxxxxxxxxxx",
					}
				case twoOrMoreCertCase:
					jsonResponse = map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id":        "fake-snis-id-01",
								"createdAt": 1425366534,
								"name":      builtinSnis[0],
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-01",
								},
							},
							{
								"id":        "fake-snis-id-02",
								"createdAt": 1438366534,
								"name":      builtinSnis[0],
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-02",
								},
							},
							{
								"id":        "fake-snis-id-03",
								"createdAt": 1425366534,
								"name":      builtinSnis[1],
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-01",
								},
							},
							{
								"id":        "fake-snis-id-04",
								"createdAt": 1438366534,
								"name":      builtinSnis[1],
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-02",
								},
							},
							// to make tests interesting, add another public cert somehow uploaded with different snis name
							{
								"id":        "fake-snis-id-05",
								"createdAt": 1438369534,
								"name":      "test.domain.com",
								"certificate": map[string]interface{}{
									"id": "fake-cert-id-03",
								},
							},
						},
						"next": "https://localhost:8000/admin/certificates?offset=xxxxxxxxxxx",
					}
				}

				if respErr := json.NewEncoder(w).Encode(jsonResponse); respErr != nil {
					t.Fatalf("Unexpected error %v", respErr)
				}
			}
		case http.MethodPost:
			if urlPath == "/admin/certificates" {
				if postCertOk {
					w.WriteHeader(http.StatusCreated)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
				var jsonResponse map[string]interface{}
				switch listCertCase {
				default:
				case emptyCertCase:
					jsonResponse = map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id":        "fake-cert-id-01",
								"createdAt": 1425366534,
								"cert":      "-----BEGIN CERTIFICATE-----...1E0MEFE=-----END CERTIFICATE-----",
								"key":       "-----BEGIN PRIVATE KEY-----...xmS8qXA==-----END PRIVATE KEY-----",
								"snis":      builtinSnis,
							},
						},
						"next": "https://localhost:8000/admin/certificates?offset=xxxxxxxxxxx",
					}
				case oneCertCase:
					jsonResponse = map[string]interface{}{
						"data": []map[string]interface{}{
							{
								"id":        "fake-cert-id-01",
								"createdAt": 1425366534,
								"cert":      "-----BEGIN CERTIFICATE-----...1E0MEFE=-----END CERTIFICATE-----",
								"key":       "-----BEGIN PRIVATE KEY-----...xmS8qXA==-----END PRIVATE KEY-----",
								"snis":      builtinSnis,
							},
							{
								"id":        "fake-cert-id-02",
								"createdAt": 1438366534,
								"cert":      "-----BEGIN CERTIFICATE-----...j5GO0XQ=-----END CERTIFICATE-----",
								"key":       "-----BEGIN PRIVATE KEY-----...XtEiYK==-----END PRIVATE KEY-----",
								"snis":      builtinSnis,
							},
						},
						"next": "https://localhost:8000/admin/certificates?offset=xxxxxxxxxxx",
					}
				}

				if respErr := json.NewEncoder(w).Encode(jsonResponse); respErr != nil {
					t.Fatalf("Unexpected error %v", respErr)
				}
			}
		case http.MethodDelete:
			if listCertCase > emptyCertCase && deleteCertOk &&
				(urlPath == "/admin/certificates/fake-cert-id-01" || urlPath == "/admin/certificates/fake-cert-id-02") {
				w.WriteHeader(http.StatusNoContent)
			} else { // other case with non-existing certificate id
				w.WriteHeader(http.StatusBadRequest)
			}
		default:
			t.Fatalf("Unexpected http method %s call to URL %s", r.Method, urlPath)
		}
	}))
}
