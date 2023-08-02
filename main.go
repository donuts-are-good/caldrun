package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const charset = "abcdefghjkmnpqrstuvwxyz23456789"

var db *sql.DB

func main() {
	dbinit()
	var err error
	db, err = dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/users", handlerUsers)
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
		db.Exec(request)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		token := r.Header.Get("User-Token")
		log.Printf("Received token: '%v'", token)
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
	err := db.QueryRow("SELECT * FROM users WHERE token = ?", token).Scan(&user.Label, &user.Token, &user.Username)
	if err != nil {
		log.Printf("Error fetching user with token: %v", err)
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

func handlerUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handlerUsersPOST(w, r)
	default:
		http.Error(w, "Invalid request method.", http.StatusNotImplemented)
	}
}

func handlerUsersPOST(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newUser struct {
		Username string `json:"username"`
	}
	err := decoder.Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user, err := dbCreateUser(db, newUser.Username)
	if err != nil {
		http.Error(w, "Error creating user: username taken or invalid", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func handlerCalendarsGET(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	calendars := dbGetCalendarsForToken(db, user)

	err := json.NewEncoder(w).Encode(calendars)
	if err != nil {
		http.Error(w, "Failed to encode calendars to JSON", http.StatusInternalServerError)
		return
	}
}

func handlerCalendarsPOST(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	var newCalendar struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&newCalendar)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	calendar, err := dbCreateCalendar(db, user, newCalendar.Name)
	if err != nil {
		http.Error(w, "Error creating calendar", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(calendar)
	if err != nil {
		http.Error(w, "Failed to encode calendar to JSON", http.StatusInternalServerError)
		return
	}
}

func handlerEventsGET(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	events := dbGetEventsForToken(db, user)

	err := json.NewEncoder(w).Encode(events)
	if err != nil {
		http.Error(w, "Failed to encode events to JSON", http.StatusInternalServerError)
		return
	}
}

func handlerEventsPOST(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(User)

	var newEvent struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		Timestamp      string   `json:"timestamp"`
		CalendarLabels []string `json:"calendar_labels"`
	}
	err := json.NewDecoder(r.Body).Decode(&newEvent)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := dbCreateEvent(db, user, newEvent.Name, newEvent.Description, newEvent.Timestamp, newEvent.CalendarLabels)
	if err != nil {
		http.Error(w, "Error creating event", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(event)
	if err != nil {
		http.Error(w, "Failed to encode event to JSON", http.StatusInternalServerError)
		return
	}
}

func dbCreateUser(db *sql.DB, username string) (User, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=?)", username).Scan(&exists)
	if err != nil {
		return User{}, err
	}
	if exists {
		return User{}, errors.New("username already exists")
	}

	thisLabel, _ := generateLabel()
	thisToken, _ := generateToken()
	user := User{
		Label:    thisLabel,
		Token:    thisToken,
		Username: username,
	}
	_, err = db.Exec("INSERT INTO users (label, token, username) VALUES (?, ?, ?)", user.Label, user.Token, user.Username)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func generateString(size int) (string, error) {
	id := make([]byte, size)
	for i := range id {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		id[i] = charset[n.Int64()]
	}
	return string(id), nil
}

func generateLabel() (string, error) {
	output, outputErr := generateString(8)
	if outputErr != nil {
		log.Println(outputErr)
	}
	return output, nil
}
func generateToken() (string, error) {
	output, outputErr := generateString(64)
	if outputErr != nil {
		log.Println(outputErr)
	}
	return output, nil
}

func dbGetCalendarsForToken(db *sql.DB, user User) []Calendar {
	calendars := []Calendar{}
	rows, err := db.Query("SELECT * FROM calendars WHERE owner_label = ?", user.Label)
	if err != nil {
		log.Printf("Error fetching calendars for user: %v", err)
		return calendars
	}
	defer rows.Close()
	for rows.Next() {
		var calendar struct {
			Calendar
			ViewUsers string `json:"view_users"`
			ModUsers  string `json:"mod_users"`
		}
		if err := rows.Scan(&calendar.Label, &calendar.OwnerLabel, &calendar.Name, &calendar.ViewUsers, &calendar.ModUsers); err != nil {
			log.Printf("Error scanning calendar row: %v", err)
			continue
		}
		viewUsers := strings.Split(calendar.ViewUsers, ",")
		modUsers := strings.Split(calendar.ModUsers, ",")
		calendars = append(calendars, Calendar{
			Label:      calendar.Label,
			OwnerLabel: calendar.OwnerLabel,
			Name:       calendar.Name,
			ViewUsers:  viewUsers,
			ModUsers:   modUsers,
		})
	}
	return calendars
}

func dbGetEventsForToken(db *sql.DB, user User) []Event {
	events := []Event{}
	rows, err := db.Query("SELECT * FROM events WHERE owner_label = ?", user.Label)
	if err != nil {
		log.Printf("Error fetching events for user: %v", err)
		return events
	}
	defer rows.Close()
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.Label, &event.OwnerLabel, &event.Name, &event.Description, &event.Timestamp, &event.CalendarLabels); err != nil {
			log.Printf("Error scanning event row: %v", err)
			continue
		}
		events = append(events, event)
	}
	return events
}

func dbCreateEvent(db *sql.DB, user User, name string, description string, timestamp string, calendarLabels []string) (Event, error) {
	log.Println("Starting dbCreateEvent")
	log.Printf("User: %v, Name: %s, Description: %s, Timestamp: %s, CalendarLabels: %v", user, name, description, timestamp, calendarLabels)

	newLabel, _ := generateLabel()
	log.Printf("Generated new label: %s", newLabel)

	_, err := db.Exec(
		"INSERT INTO events (label, owner_label, name, description, timestamp, calendar_labels) VALUES (?, ?, ?, ?, ?, ?)",
		newLabel, user.Label, name, description, timestamp, strings.Join(calendarLabels, ","),
	)

	if err != nil {
		log.Printf("Error inserting new event: %v", err)
		return Event{}, fmt.Errorf("error inserting new event: %w", err)
	}

	log.Println("Successfully created new event")

	return Event{
		Label:          newLabel,
		OwnerLabel:     user.Label,
		Name:           name,
		Description:    description,
		Timestamp:      timestamp,
		CalendarLabels: calendarLabels,
	}, nil
}

func dbCreateCalendar(db *sql.DB, user User, name string) (Calendar, error) {
	newLabel, _ := generateLabel()

	viewUsers := user.Label
	modUsers := user.Label

	_, err := db.Exec(
		"INSERT INTO calendars (label, owner_label, name, view_users, mod_users) VALUES (?, ?, ?, ?, ?)",
		newLabel, user.Label, name, viewUsers, modUsers,
	)

	if err != nil {
		return Calendar{}, fmt.Errorf("error inserting new calendar: %w", err)
	}

	return Calendar{
		Label:      newLabel,
		OwnerLabel: user.Label,
		Name:       name,
		ViewUsers:  strings.Split(viewUsers, ","),
		ModUsers:   strings.Split(modUsers, ","),
	}, nil
}
