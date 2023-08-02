package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const charset = "abcdefghjkmnpqrstuvwxyz23456789"

var db *sql.DB

type User struct {
	Label    string `json:"label,omitempty"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
}

type Calendar struct {
	Label      string   `json:"label,omitempty"`
	OwnerLabel string   `json:"owner_label,omitempty"`
	Name       string   `json:"name,omitempty"`
	ViewUsers  []string `json:"view_users,omitempty"`
	ModUsers   []string `json:"mod_users,omitempty"`
}
type Event struct {
	Label          string   `json:"label,omitempty"`
	Name           string   `json:"name,omitempty"`
	Description    string   `json:"description,omitempty"`
	Timestamp      string   `json:"timestamp,omitempty"`
	CalendarLabels []string `json:"calendar_labels,omitempty"`
}

func main() {
	dbinit()
	var err error
	db, err = dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/calendars", authMiddleware(handlerCalendars))
	http.Handle("/events", authMiddleware(handlerEvents))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		return nil, err
	}
	return db, err
}

func dbinit() {
	content, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	requests := strings.Split(string(content), ";")
	for _, request := range requests {
		_, err = db.Exec(request)
		if err != nil {
			log.Printf("Error executing sql statement %s: %v", request, err)
		}
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("User.Token")
		user, err := getUserForToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func getUserForToken(token string) (User, error) {
	var user User
	err := db.QueryRow("SELECT * FROM users WHERE token = ?", token).Scan(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func handlerCalendars(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlerCalendarsGET(w, r)
	case "POST":
		handlerCalendarsPOST(w, r)
	default:
		http.Error(w, "Invalid request method.", http.StatusNotImplemented)
	}
}

func handlerEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlerEventsGET(w, r)
	case "POST":
		handlerEventsPOST(w, r)
	default:
		http.Error(w, "Invalid request method.", http.StatusNotImplemented)
	}
}

func handlerCalendarsGET(w http.ResponseWriter, r *http.Request) {
	// Implement your logic here
}

func handlerCalendarsPOST(w http.ResponseWriter, r *http.Request) {
	// Implement your logic here
}

func handlerEventsGET(w http.ResponseWriter, r *http.Request) {
	// Implement your logic here
}

func handlerEventsPOST(w http.ResponseWriter, r *http.Request) {
	// Implement your logic here
}

func dbGetCalendarsForToken(db *sql.DB, user User) []Calendar {
	// Implement
	return []Calendar{}
}

func dbGetEventsForToken(db *sql.DB, user User) []Event {
	// Implement
	return []Event{}
}

func dbCreateEvent(db *sql.DB, user User) (Event, error) {
	// Implement
	return Event{}, errors.New("not implemented")
}

func dbCreateCalendar(db *sql.DB, user User) (Calendar, error) {
	// Implement
	return Calendar{}, errors.New("not implemented")
}
