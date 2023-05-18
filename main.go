package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = ":8080"

func main() {
	fmt.Println("Starting server")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "full Home page")
	})

	err := http.ListenAndServe(port, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server started on port: ", port)
}
