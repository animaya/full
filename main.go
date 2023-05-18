package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const port = ":8080"

func main() {

	r := mux.NewRouter()
	fmt.Println("Starting server")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "full Home page")
	})

	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Fprintf(w, "the id: %v", vars["id"])
	})

	err := http.ListenAndServe(port, r)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server started on port: ", port)
}
