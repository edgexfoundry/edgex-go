package bone

import (
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

func TestNewRoute(t *testing.T) {
	type args struct {
		url string
		h   http.Handler
	}
	tests := []struct {
		name string
		args args
		want *Route
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRoute(tt.args.url, tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_save(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
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
			r.save()
		})
	}
}

func TestRoute_Match(t *testing.T) {
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
			if got := r.Match(tt.args.req); got != tt.want {
				t.Errorf("Route.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_matchAndParse(t *testing.T) {
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
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
		want1  map[string]string
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
			got, got1 := r.matchAndParse(tt.args.req)
			if got != tt.want {
				t.Errorf("Route.matchAndParse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Route.matchAndParse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRoute_parse(t *testing.T) {
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
			if got := r.parse(tt.args.rw, tt.args.req); got != tt.want {
				t.Errorf("Route.parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_matchRawTokens(t *testing.T) {
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
		ss *[]string
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
			if got := r.matchRawTokens(tt.args.ss); got != tt.want {
				t.Errorf("Route.matchRawTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Get(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Post(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Post(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Post() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Put(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Put(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Put() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Delete(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Delete(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Head(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Head(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Head() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Patch(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Patch(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Patch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Options(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   *Route
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
			if got := r.Options(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Options() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_ServeHTTP(t *testing.T) {
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
			r.ServeHTTP(tt.args.rw, tt.args.req)
		})
	}
}
