package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// TODO Make the url validation a bit more resilient
var db *sql.DB
var config *Configuration

type TinyURL struct {
	Url string `json:"url"`
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
	r.HandleFunc("/{hash:[a-z0-9]+}", TinyUrlRedirectHandler).Methods("GET")
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
	// Converts the integer to a hash. Pretty basic but it's OK for our use
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := stmt.Exec(requrl.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Converts the integer to a hash. Pretty basic but it's OK for our use
	tinyhash := strconv.FormatInt(lastId, 36)
	tinyurl := FormatUrl(config.Proto, config.Hostname, config.Port, tinyhash)
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

	rows, err := db.Query("SELECT url FROM urls LIMIT 10 OFFSET ?", offset)
	defer rows.Close()
	// TODO We should be returning JSON and not plaintext
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		fmt.Fprintln(w, url)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
