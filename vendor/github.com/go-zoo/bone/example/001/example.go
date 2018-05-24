package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-zoo/bone"
)

var (
	mux = bone.New(Serve, Wrap)
)

func Wrap(mux *bone.Mux) *bone.Mux {
	return mux.Prefix("/api")
}

func Serve(mux *bone.Mux) *bone.Mux {
	mux.Serve = func(rw http.ResponseWriter, req *http.Request) {
		tr := time.Now()
		mux.DefaultServe(rw, req)
		fmt.Println("Serve request from", req.RemoteAddr, "in", time.Since(tr))
	}
	return mux
}

func main() {
	// Custom 404
	mux.NotFoundFunc(Handler404)
	// Handle with any http method, Handle takes http.Handler as argument.
	mux.Handle("/index", http.HandlerFunc(homeHandler))
	mux.Handle("/index/:var/info/:test", http.HandlerFunc(varHandler))
	// Get, Post etc... takes http.HandlerFunc as argument.
	mux.Post("/home", http.HandlerFunc(homeHandler))
	mux.Get("/home/:var", http.HandlerFunc(varHandler))

	mux.GetFunc("/test/*", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(req.RequestURI))
	})

	// Start Listening
	log.Fatal(mux.ListenAndServe(":8080"))
}

func homeHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("WELCOME HOME"))
}

func varHandler(rw http.ResponseWriter, req *http.Request) {
	varr := bone.GetValue(req, "var")
	test := bone.GetValue(req, "test")

	var args = struct {
		First  string
		Second string
	}{varr, test}

	if err := json.NewEncoder(rw).Encode(&args); err != nil {
		panic(err)
	}
}

func Handler404(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("These are not the droids you're looking for ..."))
}
