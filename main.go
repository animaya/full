package main

import (
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

func main() {
	//registering mux router
	r := mux.NewRouter()

	//configure db access
	//root:password@(localhost:30306)/db_test?parseTime=true

	db, err := sql.Open("mysql", os.Getenv("mysql"))

	if err == sql.ErrNoRows {
		log.Printf("No rows found in query")
	}
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.HandleFunc("/", logging(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "full Home page")
	}))

	r.HandleFunc("/form", Chain(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("form.html")
		if err != nil {
			log.Fatal(err)
		}

		if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}

		details := ContactDetail{
			Email:   r.FormValue("email"),
			Subject: r.FormValue("subject"),
			Message: r.FormValue("message"),
		}

		_ = details

		tmpl.Execute(w, struct{ Success bool }{true})
		fmt.Fprintf(w, "full Home page")
	}, Method("GET"), Logging()))

	r.HandleFunc("/layout", logging(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	r.HandleFunc("/user", logging(func(w http.ResponseWriter, r *http.Request) {
		result, err := manyRows(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "%v", result)
	}))

	fmt.Println("Starting server")
	err = http.ListenAndServe(port, r)
	if err != nil {
		fmt.Println("Server started on port: ", port)
	}

}

type ContactDetail struct {
	Email   string
	Subject string
	Message string
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

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Logging() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() { log.Println(r.URL.Path, time.Since(start)) }()
			f(w, r)
		}
	}
}

func Method(m string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			if r.Method != m {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			f(w, r)
		}
	}
}

func Chain(f http.HandlerFunc, middlwares ...Middleware) http.HandlerFunc {
	for _, m := range middlwares {
		f = m(f)
	}
	return f
}

func logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		f(w, r)
	}
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
