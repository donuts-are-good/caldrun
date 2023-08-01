package main

type User struct {
	ID       string `json:"id"`       // hash
	Token    string `json:"token"`    // access token credential, hash
	Username string `json:"username"` // human readable
}

type Event struct {
	ID          string   `json:"id"`           // hash
	Name        string   `json:"name"`         // name of this event "soccer practice"
	Description string   `json:"description"`  // description "tonight we are practicing soccer at smith field"
	Timestamp   string   `json:"timestamp"`    // datetime of this event, parse this to time.Time
	CalendarIDs []string `json:"calendar_ids"` // Calendar.id's this event belongs to
}

type Calendar struct {
	ID        string   `json:"id"`         // hash
	Name      string   `json:"name"`       // name of this calendar "mom's calendar"
	ViewUsers []string `json:"view_users"` // user.id's that can see this calendar
	ModUsers  []string `json:"mod_users"`  // user.id's that can add events to this calendar
}
