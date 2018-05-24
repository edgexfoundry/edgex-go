package bone

import (
	"net/http"
	"reflect"
	"testing"
)

func TestMux_Register(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		method  string
		path    string
		handler http.Handler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.Register(tt.args.method, tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.Register() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_GetFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.GetFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.GetFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_PostFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.PostFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.PostFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_PutFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.PutFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.PutFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_DeleteFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.DeleteFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.DeleteFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_HeadFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.HeadFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.HeadFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_PatchFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.PatchFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.PatchFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_OptionsFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
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
			if got := m.OptionsFunc(tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mux.OptionsFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMux_NotFoundFunc(t *testing.T) {
	type fields struct {
		Routes   map[string][]*Route
		prefix   string
		notFound http.Handler
		Serve    func(rw http.ResponseWriter, req *http.Request)
	}
	type args struct {
		handler http.HandlerFunc
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
			m.NotFoundFunc(tt.args.handler)
		})
	}
}
