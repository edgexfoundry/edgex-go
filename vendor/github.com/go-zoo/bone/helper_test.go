package bone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestMux_ListenAndServe(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		port string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			if err := mux.ListenAndServe(tt.args.port); (err != nil) != tt.wantErr {
				t.Errorf("Mux.ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMux_parse(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			if got := m.parse(tt.args.rw, tt.args.req); got != tt.want {
				t.Errorf("Mux.parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_staticRoute(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			if got := m.staticRoute(tt.args.rw, tt.args.req); got != tt.want {
				t.Errorf("Mux.staticRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_HandleNotFound(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
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
			m := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			m.HandleNotFound(tt.args.rw, tt.args.req)
		})
	}
}

func TestMux_validate(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			if got := m.validate(tt.args.rw, tt.args.req); got != tt.want {
				t.Errorf("Mux.validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_valid(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valid(tt.args.path); got != tt.want {
				t.Errorf("valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanURL(t *testing.T) {
	type args struct {
		url *string
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanURL(tt.args.url)
		})
	}
}

func TestGetValue(t *testing.T) {
	type args struct {
		req *http.Request
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetValue(tt.args.req, tt.args.key); got != tt.want {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_GetRequestRoute(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			if got := m.GetRequestRoute(tt.args.req); got != tt.want {
				t.Errorf("Mux.GetRequestRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetQuery(t *testing.T) {
	type args struct {
		req *http.Request
		key string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetQuery(tt.args.req, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllQueries(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAllQueries(tt.args.req); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractQueries(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 map[string][]string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := extractQueries(tt.args.req)
			if got != tt.want {
				t.Errorf("extractQueries() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("extractQueries() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMux_otherMethods(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mux{
				Routes:   tt.fields.Routes,
				prefix:   tt.fields.prefix,
				notFound: tt.fields.notFound,
				Serve:    tt.fields.Serve,
			}
			if got := m.otherMethods(tt.args.rw, tt.args.req); got != tt.want {
				t.Errorf("Mux.otherMethods() = %v, want %v", got, tt.want)
			}
		})
	}
}
