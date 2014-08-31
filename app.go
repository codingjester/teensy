package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// TODO Make the url a bit more resilient
// TODO Maybe add a caching strategy?
var db *sql.DB
var hostname = "teensy.co" // Move to a configuration or environment file

type TinyURL struct {
	Url string `json:"url"`
}

// Handles the Tiny
func TinyUrlRedirectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	hash := params["hash"]

	var url string
	var err error
	id, err := strconv.ParseInt(hash, 36, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT url FROM urls WHERE id = ?", id).Scan(&url)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func AddTinyUrlHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Authentication?
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	requrl, err := url.Parse(r.PostForm["url"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Lets insert our url into our database
	stmt, err := db.Prepare("INSERT INTO urls (url) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}

	res, err := stmt.Exec(requrl.String())
	if err != nil {
		log.Fatal(err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	tinyhash := strconv.FormatInt(lastId, 36)
	tinyurl := fmt.Sprintf("http://%s/%s", hostname, tinyhash)
	hash := TinyURL{tinyurl}
	js, err := json.Marshal(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func GetTinyUrlsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Authentication?
	offset := "0"
	params := r.URL.Query()

	_, ok := params["offset"]
	if ok {
		offset = params["offset"][0]
	}

	if _, err := strconv.Atoi(offset); err != nil {
		offset = "0"
	}

	rows, err := db.Query("SELECT url FROM urls LIMIT 10 OFFSET ?", offset)
	defer rows.Close()
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		fmt.Fprintln(w, url)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("Error querying tiny urls")
	}
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root@unix(/tmp/mysql.sock)/teensy") // Should be extracted to a setting
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

	db.SetMaxIdleConns(10)
	err = db.Ping() // Check for DB access
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

	r := mux.NewRouter()
	r.HandleFunc("/urls", AddTinyUrlHandler).Methods("POST")
	r.HandleFunc("/urls", GetTinyUrlsHandler).Methods("GET")
	r.HandleFunc("/{hash:[a-z0-9]+}", TinyUrlRedirectHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
