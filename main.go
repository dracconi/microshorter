package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

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
		if err != nil && shortened == "" {
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

		http.Redirect(w, r, url, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusNotFound)
	}

}

func main() {

	var err error
	db, err = sql.Open("sqlite3", "./main.db")
	if err != nil {
		panic(err)
	}

	rtr := mux.NewRouter()
	fmt.Printf("works!")
	rtr.HandleFunc("/{short}", shortened).Methods("GET")
	rtr.HandleFunc("/", handle).Methods("GET")
	http.Handle("/", rtr)
	log.Fatal(http.ListenAndServe(":6060", nil))

}
