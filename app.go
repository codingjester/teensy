package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB
var config *Configuration

type TinyURL struct {
	Url  string `json:"url"`
	Hash string `json:"hash"`
}

type Configuration struct {
	Hostname    string
	Proto       string
	Port        int
	Db_Type     string
	Db_Username string
	Db_Password string
	Db_Host     string
	DB          string
}

func main() {

	// Load configurations on startup
	loadConfig()

	// Setup all of our database connections
	setupDB()

	r := mux.NewRouter()
	r.HandleFunc("/urls", AddTinyUrlHandler).Methods("POST")
	r.HandleFunc("/urls", GetTinyUrlsHandler).Methods("GET")
	r.HandleFunc("/{hash:[a-z0-9]+}", TinyUrlRedirectHandler).Methods("GET") // if you want to use something with base 64, you'll need to change this regex to [a-zA-Z0-9]+
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func loadConfig() {
	file, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		log.Fatal("unable to open config: ", err)
	}

	temp := new(Configuration) // Get a pointer to an instance with new keyword
	// Unmarshal is going to decode and store into temp
	if err = json.Unmarshal(file, temp); err != nil {
		log.Println("parse config", err)
	}
	config = temp
}

// Sets up the datbase using the loaded config
// Possible improvements would be adding in other database support
func setupDB() {

	var err error
	database_url := fmt.Sprintf("%s@%s/%s", config.Db_Username, config.Db_Host, config.DB)
	db, err = sql.Open(config.Db_Type, database_url)
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

	db.SetMaxIdleConns(10)
	err = db.Ping() // Check for DB access
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}

}

// Handles the Tiny
func TinyUrlRedirectHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	hash := params["hash"]

	var url string
	var err error
	// Converts the integer to a hash, built in helpers.
	id, err := DecodeHash(hash)
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

	// the url we've parsed from the POST request
	url := r.PostForm["url"][0]

	if !ValidateURL(url) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Lets insert our url into our database
	stmt, err := db.Prepare("INSERT INTO urls (url) VALUES(?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := stmt.Exec(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encodes the integer into a hash, built in the helpers.
	tinyhash := EncodeHash(lastId)
	tinyurl := FormatUrl(config.Proto, config.Hostname, config.Port, tinyhash)
	hash := TinyURL{tinyurl, tinyhash}
	js, err := json.Marshal(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, js)
}

func GetTinyUrlsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Authentication?
	offset := "0"
	params := r.URL.Query()

	// Validates if offset exists and sets the value
	_, ok := params["offset"]
	if ok {
		offset = params["offset"][0]
	}

	// Validates if offset is capable of being an int, default
	// to zero if it can't
	if _, err := strconv.Atoi(offset); err != nil {
		offset = "0"
	}

	rows, err := db.Query("SELECT id, url FROM urls LIMIT 10 OFFSET ?", offset)
	defer rows.Close()

	urls := []TinyURL{}
	for rows.Next() {
		var id int64
		var url string
		err = rows.Scan(&id, &url)
		urls = append(urls, TinyURL{url, EncodeHash(id)})
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, js)

}
