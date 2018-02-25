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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
		?s=(original url)
		`))

	if s := r.URL.Query()["s"]; len(s) > 0 && s[0] != "" {
		var url, shortened string
		err := db.QueryRow("SELECT url, short FROM links WHERE (url='"+s[0]+"')").Scan(&url, &shortened)
		c, err := r.Cookie("auth")
		if err != nil && shortened == "" && c.Value == os.Getenv("SHORT_AUTH") {
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
		}

		w.Write([]byte(shortened))
	}
}

func shortened(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["short"]
	var url string
	err := db.QueryRow("SELECT url FROM links WHERE (short='" + short + "')").Scan(&url)
	if url != "" {
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, url, 302)
	} else {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

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
}

func teapot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
	w.Write([]byte("Oh! It's a teapot ;3"))
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
	rtr.HandleFunc("/{short}", shortened).Methods("GET")
	rtr.HandleFunc("/", handle).Methods("GET")
	http.Handle("/", rtr)
	logger.Log("Everything is up!")
	log.Fatal(http.ListenAndServe(":6060", nil))

}
