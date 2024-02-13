package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strconv"
	"encoding/json"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

)

func connect() (*sql.DB, error) {
	bin, err := ioutil.ReadFile("/run/secrets/db-password")
	if err != nil {
		return nil, err
	}
	return sql.Open("postgres", fmt.Sprintf("postgres://postgres:%s@db:5432/example?sslmode=disable", string(bin)))
}

func root(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>myapi</title>
		</head>
		<body>
			<h1>Hello there!</h1>
			<p>Welcome to my API database.</p>
			<p>Selecting tag below for using API as CRUD.</p>
			<ui>
				<li>Create 	: /create/{firstname}/{lastname}</li>
				<li>Read 	: /read</li>
				<li>Update 	: /update/{id}/{firstname}/{lastname}</li>
				<li>Delete 	: /delete/{id}</li>
			</ui>
		</body>
		</html>
	`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "404 Not Found: %s", r.URL.Path)
}

func create(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	var1 := vars["firstname"]
	var2 := vars["lastname"]

	firstname := strings.Title(var1) 
	lastname := strings.Title(var2) 

	db, err := connect()
	if err != nil {
		return
	}
	defer db.Close()

	if _, err := db.Exec("INSERT INTO human (firstname, lastname) VALUES ($1, $2);", firstname, lastname); err != nil {
		fmt.Fprintf(w, "Error creating data in the database: %s", err)
		return
	}

	fmt.Fprintf(w, "%s %s is added to human table successfully", firstname, lastname)
}

func read(w http.ResponseWriter, r *http.Request) {
	db, err := connect()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, firstname, lastname FROM human")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	var names []string
	for rows.Next() {
		var id int
		var firstname, lastname string
		err = rows.Scan(&id, &firstname, &lastname)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		fullName := fmt.Sprintf("id:%d %s %s", id, firstname, lastname)
		names = append(names, fullName)
	}
	json.NewEncoder(w).Encode(names)
}

func update(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    var1 := vars["firstname"]
	var2 := vars["lastname"]

	firstname := strings.Title(var1) 
	lastname := strings.Title(var2) 

	idValue, err := strconv.Atoi(id)
    if err != nil {
        fmt.Fprintf(w, "Error converting 'id' parameter to integer: %s", err)
        return
    }

    db, err := connect()
    if err != nil {
        fmt.Fprintf(w, "Error connecting to the database: %s", err)
        return
    }
    defer db.Close()

    if _, err := db.Exec("UPDATE human SET firstname=$1, lastname=$2 WHERE id=$3", firstname, lastname, idValue); err != nil {
        fmt.Fprintf(w, "Error updating data in the database: %s", err)
        return
    }

	fmt.Fprintf(w, "Updated blog post ID %d with the new title %s %s successfully", idValue, firstname, lastname)
}

func delete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
	

    idValue, err := strconv.Atoi(id)
    if err != nil {
        fmt.Fprintf(w, "Error converting 'id' parameter to integer: %s", err)
        return
    }

    db, err := connect()
    if err != nil {
        fmt.Fprintf(w, "Error connecting to the database: %s", err)
        return
    }
    defer db.Close()

    if _, err := db.Exec("DELETE FROM human WHERE id=$1", idValue); err != nil {
        fmt.Fprintf(w, "Error deleting data in the database: %s", err)
        return
    }

    fmt.Fprintf(w, "Deleted human at post ID %d successfully", idValue)
}

func main() {
	log.Print("Prepare db...")
	if err := prepare(); err != nil {
		log.Fatal(err)
	}

	log.Print("Listening 8000")
	r := mux.NewRouter()

	// Routing to CRUD
	r.HandleFunc("/", root)
	r.HandleFunc("/read", read)
	r.HandleFunc("/create/{firstname}/{lastname}", create)
	r.HandleFunc("/update/{id}/{firstname}/{lastname}", update)
	r.HandleFunc("/delete/{id}", delete)

	//404
	r.NotFoundHandler = http.HandlerFunc(notFound)

	log.Fatal(http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, r)))
}

func prepare() error {
	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Close()

	for i := 0; i < 60; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if _, err := db.Exec("DROP TABLE IF EXISTS human"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS human (id SERIAL, firstname VARCHAR, lastname VARCHAR)"); err != nil {
		return err
	}

	var firstnames = [4]string{"Chonakan", "Pon", "Thitiphat", "Saris"}
	var lastnames = [4]string{"Chumtap", "Morin", "Preededilok", "Buaiam"}

	for i := 0; i < 4; i++ {
		if _, err := db.Exec("INSERT INTO human (firstname, lastname) VALUES ($1, $2);", firstnames[i], lastnames[i]); err != nil {
			return err
		}
	}
	return nil
}