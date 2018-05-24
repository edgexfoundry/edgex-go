package bone

import (
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

func TestGetAllValues(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAllValues(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_serveMatchedRequest(t *testing.T) {
	type fields struct {
		Path    string
		Method  string
		Size    int
		Atts    int
		wildPos int
		Token   Token
		Pattern map[int]string
		Compile map[int]*regexp.Regexp
		Tag     map[int]string
		Handler http.Handler
	}
	type args struct {
		rw   http.ResponseWriter
		req  *http.Request
		vars map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Route{
				Path:    tt.fields.Path,
				Method:  tt.fields.Method,
				Size:    tt.fields.Size,
				Atts:    tt.fields.Atts,
				wildPos: tt.fields.wildPos,
				Token:   tt.fields.Token,
				Pattern: tt.fields.Pattern,
				Compile: tt.fields.Compile,
				Tag:     tt.fields.Tag,
				Handler: tt.fields.Handler,
			}
			r.serveMatchedRequest(tt.args.rw, tt.args.req, tt.args.vars)
		})
	}
}
