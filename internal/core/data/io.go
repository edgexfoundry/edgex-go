package data

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/OneOfOne/xxhash"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/ugorji/go/codec"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
)

const (
	checksumContextKey = "payload-checksum"
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
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return event, err
	}
	err = json.Unmarshal(bytes, &event)
	if err != nil {
		return event, err
	}
	event.Bytes = bytes
	return event, nil
}

// cborReader handles unmarshaling of a CBOR request body payload
type cborReader struct{}

// Read reads and converts the request's CBOR event data into an Event struct
func (cborReader) Read(reader io.Reader, ctx *context.Context) (models.Event, error) {
	c := context.WithValue(*ctx, clients.ContentType, clients.ContentTypeCBOR)
	event := models.Event{}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return event, err
	}

	x := codec.CborHandle{}
	err = codec.NewDecoderBytes(bytes, &x).Decode(&event)
	if err != nil {
		return event, err
	}

	switch Configuration.Writable.ChecksumAlgo {
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
func NewRequestReader(request *http.Request) EventReader {
	contentType := request.Header.Get(clients.ContentType)

	switch contentType {
	case clients.ContentTypeCBOR:
		return cborReader{}
	default:
		return jsonReader{}
	}
}
