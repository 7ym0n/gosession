package main

import (
	"fmt"
	s "github.com/wackonline/gosession"
	"net/http"
)

var provider *s.Adapter
var session s.SessionStore

type MyMux struct {
}

func (p *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	provider, _ = s.Bootstrap("file", `{"cookieName":"gosessionid","Gctime":3600,"ProviderConfig":"./tmp"}`)
	// fmt.Println("...2")

	session, _ = provider.StartSession(w, r)
	if r.URL.Path == "/" {
		sayhelloName(w, r)
		return
	}
	if r.URL.Path == "/hi" {
		hi(w, r)
		return
	}
	if r.URL.Path == "/hello" {
		hello(w, r)
		return
	}

	http.NotFound(w, r)
	return
}

func hello(w http.ResponseWriter, r *http.Request) {
	session.Delete("hello")
	fmt.Fprintf(w, "format, ...")
}

func hi(w http.ResponseWriter, r *http.Request) {
	// var c map[interface{}]interface{}
	nh1 := session.Get("hello")
	fmt.Fprintln(w, nh1)
	// c = session.All()

	fmt.Println(session.All())
	fmt.Fprintln(w, "say hi!!!")
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {

	session.Set("hello", "hello world!...")

	fmt.Fprintln(w, "================")
}
func main() {

	mux := &MyMux{}
	http.ListenAndServe(":8080", mux)
}
