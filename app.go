package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// TODO Make the url a bit more resilient
// TODO Make the TinyURL Struct a bit cleaner
// TODO Add stats tracking
// TODO Maybe add a caching strategy?
var db *sql.DB

type TinyURL struct {
	Hash string `json:"hash"`
}

// Handles the Tiny
func TinyUrlRedirectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	hash := params["hash"]

	var url string
	var err error
	err = db.QueryRow("SELECT url FROM urls WHERE id = ?", hash).Scan(&url)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, url, http.StatusMovedPermanently)
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

	// Lets insert our url into our database
	stmt, err := db.Prepare("INSERT INTO urls (url) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(requrl.String())
	if err != nil {
		log.Fatal(err)
	}

	/*lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}*/

	hash := TinyURL{"testing"}
	js, err := json.Marshal(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func GetTinyUrlsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Allow for pagination?
	rows, err := db.Query("SELECT url FROM urls LIMIT 10")
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
	db, err = sql.Open("mysql", "root@unix(/tmp/mysql.sock)/teensy")
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

	db.SetMaxIdleConns(10)
	err = db.Ping() // This DOES open a connection if necessary. This makes sure the database is accessible
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
