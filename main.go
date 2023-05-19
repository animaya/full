package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

const port = ":8080"

var ctx context.Context

func main() {
	//registering mux router
	r := mux.NewRouter()

	//configure db access
	//root:password@(localhost:30306)/db_test?parseTime=true
	//os.Getenv("mysql")
	db, err := sql.Open("mysql", os.Getenv("mysql"))

	if err == sql.ErrNoRows {
		log.Printf("No rows found in query")
	}
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "full Home page")
	})

	r.HandleFunc("/layout", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("layout.html")
		if err != nil {
			log.Fatal(err)
		}

		data := TodoPageData{
			PageTitle: "My TODO List",
			Todos: []Todo{
				{Title: "Task1", Done: false},
				{Title: "Task2", Done: true},
				{Title: "Task3", Done: true},
			},
		}
		tmpl.Execute(w, data)
		fmt.Fprintf(w, "full Home page")
	})

	// r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	var (
	// 		id        int
	// 		username  string
	// 		password  string
	// 		createdAt time.Time
	// 	)
	// 	query := `SELECT id, username, password, created_at FROM users WHERE id = ?`
	// 	vars := mux.Vars(r)
	// 	err := db.QueryRow(query, vars["id"]).Scan(&id, &username, &password, &createdAt)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Fprintf(w, "the id: %v, %v, %v, %v\n", id, username, password, createdAt)
	// })

	r.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		result, err := manyRows(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "%v", result)
	})

	fmt.Println("Starting server")
	err = http.ListenAndServe(port, r)
	if err != nil {
		fmt.Println("Server started on port: ", port)
	}

}

type Todo struct {
	Title string
	Done  bool
}

type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

type User struct {
	id        int
	username  string
	password  string
	createdAt time.Time
}

type City struct {
	Id         int
	Name       string
	Population int
}

func manyRows(db *sql.DB) (usr []City, err error) {

	var cities []City

	rows, err := db.Query(`SELECT * FROM cities`) // check err

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var city City
		err := rows.Scan(&city.Id, &city.Name, &city.Population)

		if err != nil {
			return cities, err
		}
		cities = append(cities, city)

	}
	if err = rows.Err(); err != nil {
		return cities, err
	}

	return cities, nil
}
