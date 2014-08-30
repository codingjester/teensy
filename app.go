package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

// TODO Make the url a bit more resilient
// TODO Make the TinyURL Struct a bit cleaner
// TODO Use a persisent data store, abstracted preferably
// TODO Add stats tracking
// TODO Maybe add a caching strategy?

// Using an in memory map
var m = map[string]string{
	"testing":  "http://google.com",
	"testing2": "http://tumblr.com",
}

type TinyURL struct {
	Hash string `json:"hash"`
}

// Handles the Tiny
func TinyUrlRedirectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	hash := params["hash"]
	url, ok := m[hash]
	if ok {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	}
}

func AddTinyUrlHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	// Make this a bit more robust bc this is sad
	requrl, err := url.Parse(r.PostForm["url"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO insert into a persisent data store and generate the tinyurl
	m["blahdy"] = requrl.String()
	hash := TinyURL{"blahdy"}
	js, err := json.Marshal(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func GetTinyUrlsHandler(w http.ResponseWriter, r *http.Request) {
	for k, v := range m {
		line := fmt.Sprintf("%s => %s", k, v)
		fmt.Fprintln(w, line)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/urls", AddTinyUrlHandler).Methods("POST")
	r.HandleFunc("/urls", GetTinyUrlsHandler).Methods("GET")
	r.HandleFunc("/{hash:[a-z0-9]+}", TinyUrlRedirectHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
