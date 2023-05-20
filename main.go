package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

const port = ":8080"

var (
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

	r.HandleFunc("/hash/{password}", logging(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		password := vars["password"]
		hash, _ := HashPassword(password) // ignore error for the sake of simplicity
		fmt.Fprintf(w, "hashing password\n")

		fmt.Fprintf(w, password+"\n")
		fmt.Fprintf(w, hash+"\n")

		match := CheckPasswordHash(password, hash)
		fmt.Fprintf(w, "checking hash\n")
		if match {
			fmt.Fprintf(w, "true")
		}

	}))

	r.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)

		for {
			msgType, msg, err := conn.ReadMessage()

			if err != nil {
				return
			}

			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}

	})

	r.HandleFunc("/websock", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})

	// r.HandleFunc("/form", Chain(func(w http.ResponseWriter, r *http.Request) {
	// 	tmpl, err := template.ParseFiles("form.html")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	if r.Method != http.MethodPost {
	// 		tmpl.Execute(w, nil)
	// 		return
	// 	}

	// 	details := ContactDetail{
	// 		Email:   r.FormValue("email"),
	// 		Subject: r.FormValue("subject"),
	// 		Message: r.FormValue("message"),
	// 	}

	// 	_ = details

	// 	tmpl.Execute(w, struct{ Success bool }{true})
	// 	fmt.Fprintf(w, "full Home page")
	// }, Method("GET"), Logging()))

	// r.HandleFunc("/layout", logging(func(w http.ResponseWriter, r *http.Request) {
	// 	tmpl, err := template.ParseFiles("layout.html")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	data := TodoPageData{
	// 		PageTitle: "My TODO List",
	// 		Todos: []Todo{
	// 			{Title: "Task1", Done: false},
	// 			{Title: "Task2", Done: true},
	// 			{Title: "Task3", Done: true},
	// 		},
	// 	}
	// 	tmpl.Execute(w, data)
	// 	fmt.Fprintf(w, "full Home page")
	// }))

	// r.HandleFunc("/user", logging(func(w http.ResponseWriter, r *http.Request) {
	// 	result, err := manyRows(db)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Fprintf(w, "%v", result)
	// }))

	fmt.Println("Starting server")

	// r.HandleFunc("/secret", secret)
	// r.HandleFunc("/login", login)
	// r.HandleFunc("/logout", logout)

	r.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
		var user Usero
		json.NewDecoder(r.Body).Decode(&user)
		fmt.Fprintf(w, "%s %s is %d years old!", user.Firstname, user.Lastname, user.Age)
	})

	r.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		peter := Usero{
			Firstname: "Peter",
			Lastname:  "Abramov",
			Age:       32,
		}
		json.NewEncoder(w).Encode(peter)
	})

	err = http.ListenAndServe(port, r)
	if err != nil {
		fmt.Println("Server started on port: ", port)
	}

}

type Usero struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	fmt.Fprintln(w, "the cake is a lie")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = true
	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
}

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
