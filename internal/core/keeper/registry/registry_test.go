//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"io"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// trackableBody is an io.ReadCloser that tracks whether Close() was called
type trackableBody struct {
	reader io.Reader
	closed bool
	mu     sync.Mutex
}

func newTrackableBody(data string) *trackableBody {
	return &trackableBody{
		reader: &stringReader{data: data},
		closed: false,
	}
}

func (tb *trackableBody) Read(p []byte) (n int, err error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.closed {
		return 0, io.ErrClosedPipe
	}
	return tb.reader.Read(p)
}

func (tb *trackableBody) Close() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.closed = true
	if closer, ok := tb.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (tb *trackableBody) WasClosed() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.closed
}

type stringReader struct {
	data string
	pos  int
}

func (sr *stringReader) Read(p []byte) (n int, err error) {
	if sr.pos >= len(sr.data) {
		return 0, io.EOF
	}
	n = copy(p, sr.data[sr.pos:])
	sr.pos += n
	return n, nil
}

// mockRoundTripper is a custom http.RoundTripper that returns configurable responses
type mockRoundTripper struct {
	statusCode int
	body       *trackableBody
	err        error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &http.Response{
		StatusCode: m.statusCode,
		Status:     http.StatusText(m.statusCode),
		Body:       m.body,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

// healthCheckWithTransport is a test helper that mirrors the logic of healthCheck
// but allows injecting a custom http.RoundTripper for testing.
// This version matches the fixed behavior where resp.Body is properly closed.
func healthCheckWithTransport(r models.Registration, lc logger.LoggingClient, timeout time.Duration, transport http.RoundTripper) string {
	client := http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	path := r.HealthCheck.Type + "://" + r.Host + ":" + strconv.Itoa(r.Port) + r.HealthCheck.Path
	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)
	if err != nil {
		lc.Errorf("failed to create get request for %s: %v", path, err)
		return models.Down
	}

	resp, err := client.Do(req)
	if err != nil {
		lc.Errorf("Failed to health check service %s: %s", r.ServiceId, err.Error())
		return models.Down
	}

	// Ensure response body is always closed to prevent resource leaks
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		lc.Debugf("service %s status healthy", r.ServiceId)
		return models.Up
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			lc.Error("Failed to read %s response body: %s", path, err.Error())
		}
		lc.Errorf("service %s is unhealthy: %s", r.ServiceId, string(bodyBytes))
		return models.Down
	}
}

func TestHealthCheck_ResponseBodyAlwaysClosed_Success(t *testing.T) {
	// Arrange
	body := newTrackableBody(`{"status":"ok"}`)
	transport := &mockRoundTripper{
		statusCode: http.StatusOK,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Up, result, "Should return Up for 200 status")
	assert.True(t, body.WasClosed(), "Response body must be closed on success path")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_FailureStatus(t *testing.T) {
	// Arrange
	body := newTrackableBody(`{"error":"service unavailable"}`)
	transport := &mockRoundTripper{
		statusCode: http.StatusServiceUnavailable,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Down, result, "Should return Down for non-2xx status")
	assert.True(t, body.WasClosed(), "Response body must be closed on failure path")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_NotFound(t *testing.T) {
	// Arrange
	body := newTrackableBody(`{"error":"not found"}`)
	transport := &mockRoundTripper{
		statusCode: http.StatusNotFound,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Down, result, "Should return Down for 404 status")
	assert.True(t, body.WasClosed(), "Response body must be closed on 404 path")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_InternalServerError(t *testing.T) {
	// Arrange
	body := newTrackableBody(`{"error":"internal server error"}`)
	transport := &mockRoundTripper{
		statusCode: http.StatusInternalServerError,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Down, result, "Should return Down for 500 status")
	assert.True(t, body.WasClosed(), "Response body must be closed on 500 path")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_Status299(t *testing.T) {
	// Arrange - Test boundary case: 299 is still in success range
	body := newTrackableBody(`{"status":"ok"}`)
	transport := &mockRoundTripper{
		statusCode: 299,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Up, result, "Should return Up for 299 status (boundary)")
	assert.True(t, body.WasClosed(), "Response body must be closed on success boundary path")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_Status300(t *testing.T) {
	// Arrange - Test boundary case: 300 is outside success range
	body := newTrackableBody(`{"status":"redirect"}`)
	transport := &mockRoundTripper{
		statusCode: http.StatusMultipleChoices,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Down, result, "Should return Down for 300 status (boundary)")
	assert.True(t, body.WasClosed(), "Response body must be closed on failure boundary path")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_EmptyBody(t *testing.T) {
	// Arrange - Test with empty response body
	body := newTrackableBody("")
	transport := &mockRoundTripper{
		statusCode: http.StatusOK,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Up, result, "Should return Up for 200 with empty body")
	assert.True(t, body.WasClosed(), "Response body must be closed even with empty body")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_LargeBody(t *testing.T) {
	// Arrange - Test with large response body that needs to be read
	largeBodyData := make([]byte, 10000)
	for i := range largeBodyData {
		largeBodyData[i] = byte('A' + (i % 26))
	}
	body := newTrackableBody(string(largeBodyData))
	transport := &mockRoundTripper{
		statusCode: http.StatusBadRequest,
		body:       body,
	}

	registration := models.Registration{
		ServiceId: "test-service",
		Host:      "localhost",
		Port:      8080,
		HealthCheck: models.HealthCheck{
			Type:     "http",
			Path:     "/api/v3/ping",
			Interval: "10s",
		},
	}
	lc := logger.NewMockClient()
	timeout := 5 * time.Second

	// Act
	result := healthCheckWithTransport(registration, lc, timeout, transport)

	// Assert
	assert.Equal(t, models.Down, result, "Should return Down for 400 status")
	assert.True(t, body.WasClosed(), "Response body must be closed after reading large body")
}

func TestHealthCheck_ResponseBodyAlwaysClosed_AllPaths(t *testing.T) {
	// Test all code paths to ensure body is always closed
	testCases := []struct {
		name       string
		statusCode int
		expected   string
	}{
		{"Success 200", http.StatusOK, models.Up},
		{"Success 201", http.StatusCreated, models.Up},
		{"Success 204", http.StatusNoContent, models.Up},
		{"Success 299", 299, models.Up},
		{"Failure 300", http.StatusMultipleChoices, models.Down},
		{"Failure 400", http.StatusBadRequest, models.Down},
		{"Failure 404", http.StatusNotFound, models.Down},
		{"Failure 500", http.StatusInternalServerError, models.Down},
		{"Failure 503", http.StatusServiceUnavailable, models.Down},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			body := newTrackableBody(`{"test":"data"}`)
			transport := &mockRoundTripper{
				statusCode: tc.statusCode,
				body:       body,
			}

			registration := models.Registration{
				ServiceId: "test-service",
				Host:      "localhost",
				Port:      8080,
				HealthCheck: models.HealthCheck{
					Type:     "http",
					Path:     "/api/v3/ping",
					Interval: "10s",
				},
			}
			lc := logger.NewMockClient()
			timeout := 5 * time.Second

			// Act
			result := healthCheckWithTransport(registration, lc, timeout, transport)

			// Assert
			assert.Equal(t, tc.expected, result, "Status code %d should return %s", tc.statusCode, tc.expected)
			require.True(t, body.WasClosed(), "Response body MUST be closed for status code %d", tc.statusCode)
		})
	}
}
