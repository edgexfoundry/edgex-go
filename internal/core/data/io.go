package data

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/OneOfOne/xxhash"
	"github.com/fxamacker/cbor/v2"
)

const (
	checksumContextKey = "payload-checksum"
	maxEventSize       = int64(25 * 1e6) // 25 MB
)

// ErrUnsupportedContentType an error used when a request is received with an unsupported content type
type ErrUnsupportedContentType struct {
	ContentType string
}

func (e ErrUnsupportedContentType) Error() string {
	return "Unsupported content type: '" + e.ContentType + "'"
}

// EventReader unmarshals a request body into an Event type
type EventReader interface {
	Read(reader io.Reader, ctx *context.Context) (models.Event, error)
}

// jsonReader handles unmarshaling of a JSON request body payload
type jsonReader struct{}

// Read reads and converts the request's JSON event data into an Event struct
func (jsonReader) Read(reader io.Reader, ctx *context.Context) (models.Event, error) {
	c := context.WithValue(*ctx, clients.ContentType, clients.ContentTypeJSON)
	*ctx = c

	event := models.Event{}
	err := json.NewDecoder(reader).Decode(&event)
	if err != nil {
		return event, err
	}

	return event, nil
}

// NewJsonReader creates a new instance of cborReader.
func NewJsonReader() jsonReader {
	return jsonReader{}
}

// cborReader handles unmarshaling of a CBOR request body payload
type cborReader struct {
	configuration *config.ConfigurationStruct
}

// NewCborReader creates a new instance of cborReader.
func NewCborReader(configuration *config.ConfigurationStruct) cborReader {
	return cborReader{configuration: configuration}
}

// Read reads and converts the request's CBOR event data into an Event struct
func (cr cborReader) Read(reader io.Reader, ctx *context.Context) (models.Event, error) {
	c := context.WithValue(*ctx, clients.ContentType, clients.ContentTypeCBOR)
	event := models.Event{}
	bytes, err := ioutil.ReadAll(io.LimitReader(reader, maxEventSize))
	if err != nil {
		return event, err
	}

	err = cbor.Unmarshal(bytes, &event)
	if err != nil {
		return event, err
	}

	switch cr.configuration.Writable.ChecksumAlgo {
	case ChecksumAlgoxxHash:
		event.Checksum = fmt.Sprintf("%x", xxhash.Checksum64(bytes))
	default:
		event.Checksum = fmt.Sprintf("%x", md5.Sum(bytes))
	}
	c = context.WithValue(c, checksumContextKey, event.Checksum)
	*ctx = c
	event.Bytes = bytes

	return event, nil
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewRequestReader(request *http.Request, configuration *config.ConfigurationStruct) EventReader {
	contentType := request.Header.Get(clients.ContentType)

	switch strings.ToLower(contentType) {
	case clients.ContentTypeCBOR:
		return NewCborReader(configuration)
	default:
		return NewJsonReader()
	}
}
