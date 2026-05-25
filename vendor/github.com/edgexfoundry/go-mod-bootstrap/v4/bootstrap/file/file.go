package file

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

func Load(path string, provider interfaces.SecretProvider, lc logger.LoggingClient) ([]byte, error) {
	var fileBytes []byte
	var err error

	parsedUrl, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("could not parse file path: %v", err)
	}

	if parsedUrl.Scheme == config.DefaultHttpProtocol || parsedUrl.Scheme == config.HttpsProtocol {
		client := &http.Client{
			Timeout: environment.GetURIRequestTimeout(lc),
		}
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to create new request for remote file: %s: %v", parsedUrl.Redacted(), err)
		}

		// Get httpheader secret
		params := parsedUrl.Query()
		edgexSecretName := params.Get("edgexSecretName")
		if edgexSecretName != "" {
			secrets, err := provider.GetSecret(edgexSecretName)
			if err != nil {
				return nil, err
			}

			// Set request header
			if len(secrets) > 0 && secrets["type"] == "httpheader" {
				if secrets["headername"] != "" && secrets["headercontents"] != "" {
					req.Header.Add(secrets["headername"], secrets["headercontents"])
				} else {
					return nil, fmt.Errorf("secret headername and headercontents can not be empty")
				}
			} else {
				return nil, fmt.Errorf("secret type is not httpheader")
			}
		}

		// Run request
		resp, err := client.Do(req)

		if err != nil {
			return nil, fmt.Errorf("could not get remote file: %s: %v", parsedUrl.Redacted(), err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				lc.Errorf("error closing response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("invalid status code %d loading remote file: %s", resp.StatusCode, parsedUrl.Redacted())
		}

		fileBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read remote file: %s: %v", parsedUrl.Redacted(), err)
		}
	} else {
		fileBytes, err = os.ReadFile(path) // #nosec G304 -- path is controlled and safe to be read
		if err != nil {
			return nil, fmt.Errorf("could not read file %s: %v", path, err)
		}
	}

	return fileBytes, nil
}
