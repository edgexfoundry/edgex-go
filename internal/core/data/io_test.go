package data

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var TestEvent = contract.Event{
	ID:       "TestEvent",
	Pushed:   0623,
	Device:   "TestDevice",
	Created:  1553247000,
	Modified: 1573900200,
	Origin:   9,
	Readings: buildReadings(),
}

func TestNewReader(t *testing.T) {
	type args struct {
		contentType string
	}
	tests := []struct {
		name    string
		args    args
		want    EventReader
		wantErr bool
	}{
		{
			name: "Get Json Reader",
			args: args{
				contentType: clients.ContentTypeJSON,
			},
			want: jsonReader{},
		},
		{
			name: "Get Cbor Reader",
			args: args{
				contentType: clients.ContentTypeCBOR,
			},
			want: cborReader{},
		},
		{
			name: "Get Reader for unknown type",
			args: args{
				contentType: "unkown-type",
			},
			want: jsonReader{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := newRequestWithContentType(tt.args.contentType)
			got := NewRequestReader(request, &config.ConfigurationStruct{})

			eet := reflect.TypeOf(tt.want)
			aet := reflect.TypeOf(got)
			if !aet.AssignableTo(eet) {
				t.Errorf("Expected reader of type %v, but got an reader of type %v", eet, aet)
			}
		})
	}
}

func TestJsonSerialization(t *testing.T) {
	event := models.Event{Event: TestEvent}

	data, _ := event.MarshalJSON()

	r := ioutil.NopCloser(bytes.NewBuffer(data))

	ctx := context.Background()
	jsonReader := NewJsonReader()
	_, err := jsonReader.Read(r, &ctx)

	if err != nil {
		t.Errorf("Should not encounter an error")
	}

}

func TestCborSerialization(t *testing.T) {
	reset()
	event := models.Event{Event: TestEvent}
	data := event.CBOR()
	r := ioutil.NopCloser(bytes.NewBuffer(data))

	cborReader := NewCborReader(&config.ConfigurationStruct{})

	ctx := context.Background()
	result, err := cborReader.Read(r, &ctx)

	if err != nil {
		t.Errorf("Should not encounter an error")
	}
	if !reflect.DeepEqual(*result.ToContract(), TestEvent) {
		t.Errorf("TestCborSerialization() = %v, want %v", result, TestEvent)
	}

}

func newRequestWithContentType(contentType string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Test body"))
	req.Header.Set(clients.ContentType, contentType)
	return req
}
