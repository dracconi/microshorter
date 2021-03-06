package main

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dracconi/microshorter/logger"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(
		`
		GET s=URL - shorten
		`))

	if s := r.URL.Query()["s"]; len(s) > 0 && s[0] != "" {
		var url, shortened string
		logger.Log("(301) Trying to shorten " + s[0])
		errd := db.QueryRow("SELECT url, short FROM links WHERE (url='"+s[0]+"')").Scan(&url, &shortened)
		c, err := r.Cookie("auth")
		if errd != nil {
			logger.Log(errd.Error())
		}
		if err != nil {
			logger.Log(err.Error())
		}
		if shortened == "" && c.Value == os.Getenv("SHORT_AUTH") {
			logger.Log(r.RemoteAddr + " authenticated properly")
			query, err := db.Prepare("INSERT INTO links(url, short) VALUES (?,?)")
			if err != nil {
				panic(err)
			}
			rand.Seed(time.Now().UnixNano())
			shortened = randStringBytes(4)
			_, err = query.Exec(s[0], shortened)
			if err != nil {
				panic(err)
			}
			logger.Log(s[0] + " -> " + shortened)
		}

		w.Write([]byte(shortened))
	}
	defer r.Body.Close()
}

func shortened(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]
	var url string
	err := db.QueryRow("SELECT url FROM links WHERE (short='" + short + "')").Scan(&url)
	if url != "" {
		if err != nil {
			panic(err)
		}
		logger.Log("(302) " + short + " -> " + url + " | " + r.RemoteAddr)
		http.Redirect(w, r, url, 302)
	} else {
		logger.Log("(404)" + short + " | " + r.RemoteAddr)
		http.Redirect(w, r, "/", http.StatusNotFound)
	}
	defer r.Body.Close()
}

// dumplings
func dumplinks(w http.ResponseWriter, r *http.Request) {
	resp, err := db.Query("SELECT url, short FROM links")
	if err != nil {
		panic(err)
	}
	var url, short string
	w.Write([]byte("url | shortened\n----+----------\n"))
	for resp.Next() {
		resp.Scan(&url, &short)
		w.Write([]byte(url + " | " + short + "\n"))
	}
	defer r.Body.Close()
}

func teapot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte("Oh! It's a teapot ;3"))
	defer r.Body.Close()
}

func authenticate(w http.ResponseWriter, r *http.Request) {

}

func main() {

	var err error
	db, err = sql.Open("sqlite3", "./main.db")
	if err != nil {
		panic(err)
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/links", dumplinks).Methods("GET")
	rtr.HandleFunc("/teapot", teapot).Methods("GET")
	rtr.HandleFunc("/{short:[a-zA-Z0-9]{4}}", shortened).Methods("GET")
	rtr.HandleFunc("/a", authenticate).Methods("GET")
	rtr.HandleFunc("/", handle).Methods("GET")
	http.Handle("/", rtr)
	logger.Log("Everything is up!")
	log.Fatal(http.ListenAndServe(":6060", nil))

}
