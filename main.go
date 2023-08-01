package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const charset = "abcdefghjkmnpqrstuvwxyz23456789"

func dbInit() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./calendar.db")
	if err != nil {
		return nil, err
	}

	// Read the SQL from schema.sql file
	sqlFile := "schema.sql"
	sqlContent, err := os.ReadFile(sqlFile)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to read SQL file: %v", err)
	}

	// Execute the SQL commands
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to execute SQL commands: %v", err)
	}

	return db, nil
}

func api(db *sql.DB) {
	http.Handle("/calendars", AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleCalendars(db, w, r)
	})))

	http.Handle("/events", AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleEvents(db, w, r)
	})))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handleHealth(db, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func prompt(db *sql.DB) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter a command: ")
	for scanner.Scan() {
		command := scanner.Text()
		switch strings.ToLower(command) {
		case "newuser":
			fmt.Print("Enter username: ")
			scanner.Scan()
			username := scanner.Text()

			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM Users WHERE Username = ?", username).Scan(&count)
			if err != nil {
				log.Fatal(err)
			}
			if count > 0 {
				fmt.Println("Username is taken. Please try another one.")
				continue
			}

			h := sha256.New()
			h.Write([]byte(username))
			token := hex.EncodeToString(h.Sum(nil))

			id, err := generateID()
			if err != nil {
				log.Fatal(err)
			}

			_, err = db.Exec("INSERT INTO Users (ID, Token, Username) VALUES (?, ?, ?)", id, token, username)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("New user created! Username: %s, Token: %s, ID: %s\n", username, token, id)

		default:
			fmt.Println("Unknown command")
		}
		fmt.Print("Enter a command: ")
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
}

func main() {
	db, err := dbInit()
	if err != nil {
		log.Fatal(err)
	}

	go api(db)

	prompt(db)
}

func handleHealth(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes (adjust the max-age value as needed)

	healthCheck, err := getHealthCheck(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthCheck)
}

func getHealthCheck(db *sql.DB) (HealthCheck, error) {
	var users, calendars, events int

	err := db.QueryRow("SELECT COUNT(*) FROM Users").Scan(&users)
	if err != nil {
		return HealthCheck{}, err
	}

	err = db.QueryRow("SELECT COUNT(*) FROM Calendars").Scan(&calendars)
	if err != nil {
		return HealthCheck{}, err
	}

	err = db.QueryRow("SELECT COUNT(*) FROM Events").Scan(&events)
	if err != nil {
		return HealthCheck{}, err
	}

	return HealthCheck{
		Users:     users,
		Calendars: calendars,
		Events:    events,
		Time:      time.Now(),
	}, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		id := r.URL.Query().Get("id")

		access, err := userHasAccess(token, id)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		if !access {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type HealthCheck struct {
	Users     int       `json:"users"`
	Calendars int       `json:"calendars"`
	Events    int       `json:"events"`
	Time      time.Time `json:"time"`
}

func handleCalendars(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization") // assuming the token is sent in the Authorization header
	calendarID := r.URL.Query().Get("id")  // assuming the calendar ID is sent as a query parameter

	switch r.Method {
	case http.MethodGet:
		// Retrieve all calendars from the database.
		calendars, err := getAllCalendarsForToken(db, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Encode calendars to JSON and send in the response.
		json.NewEncoder(w).Encode(calendars)
	case http.MethodPost:
		hasAccess, err := userHasAccess(token, calendarID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !hasAccess {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var cal Calendar
		err = json.NewDecoder(r.Body).Decode(&cal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Generate a unique ID for the new calendar.
		cal.ID, err = generateID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Add the new calendar to the database.
		err = createCalendar(cal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Return the newly created calendar.
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(cal)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleEvents(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization") // assuming the token is sent in the Authorization header
	eventID := r.URL.Query().Get("id")     // assuming the event ID is sent as a query parameter

	switch r.Method {
	case http.MethodGet:
		// Retrieve all events from the database.
		events, err := getAllEventsForToken(db, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Encode events to JSON and send in the response.
		json.NewEncoder(w).Encode(events)
	case http.MethodPost:
		hasAccess, err := userHasAccess(token, eventID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !hasAccess {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var ev Event
		err = json.NewDecoder(r.Body).Decode(&ev)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Generate a unique ID for the new event.
		ev.ID, err = generateID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Add the new event to the database.
		err = createEvent(db, ev)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Return the newly created event.
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ev)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func getAllEventsForToken(db *sql.DB, token string) ([]Event, error) {
	rows, err := db.Query("SELECT events.id, events.name, events.description, events.timestamp, events.userID FROM events JOIN EventCalendars ON events.id = EventCalendars.EventID WHERE events.userID = (SELECT id FROM users WHERE token = ?)", token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var ev Event
		err = rows.Scan(&ev.ID, &ev.Name, &ev.Description, &ev.Timestamp, &ev.ID)
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func createEvent(db *sql.DB, ev Event) error {
	_, err := db.Exec(
		"INSERT INTO events (id, name, description, timestamp, userID) VALUES (?, ?, ?, ?, ?)",
		ev.ID, ev.Name, ev.Description, ev.Timestamp, ev.ID,
	)
	return err
}
func getAllCalendarsForToken(db *sql.DB, token string) ([]Calendar, error) {
	rows, err := db.Query("SELECT calendars.id, calendars.name FROM calendars JOIN CalendarUsers ON calendars.id = CalendarUsers.CalendarID WHERE CalendarUsers.userID = (SELECT id FROM users WHERE token = ?)", token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calendars []Calendar
	for rows.Next() {
		var cal Calendar
		err = rows.Scan(&cal.ID, &cal.Name)
		if err != nil {
			return nil, err
		}
		calendars = append(calendars, cal)
	}
	return calendars, rows.Err()
}

func userHasAccess(token string, id string) (bool, error) {
	var count int
	// Check CalendarUsers table
	err := db.QueryRow("SELECT COUNT(*) FROM CalendarUsers WHERE userID = (SELECT id FROM users WHERE token = ?) AND calendarID = ?", token, id).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check EventCalendars table
	err = db.QueryRow("SELECT COUNT(*) FROM EventCalendars WHERE EventID = (SELECT ID FROM events WHERE userID = (SELECT id FROM users WHERE token = ?)) AND CalendarID = ?", token, id).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// User doesn't have access
	return false, nil
}

func createCalendar(cal Calendar) error {
	_, err := db.Exec(
		"INSERT INTO calendars (id, name) VALUES (?, ?)",
		cal.ID, cal.Name, // adjust this line to match your actual database structure
	)
	return err
}

func generateID() (string, error) {
	id := make([]byte, 8)
	for i := range id {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		id[i] = charset[n.Int64()]
	}
	return string(id), nil
}
