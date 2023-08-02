package main

import "time"

// type User struct {
// 	Label    string `json:"label,omitempty"`
// 	Token    string `json:"token,omitempty"`
// 	Username string `json:"username,omitempty"`
// }

// type Calendar struct {
// 	Label      string   `json:"label,omitempty"`
// 	OwnerLabel string   `json:"owner_label,omitempty"`
// 	Name       string   `json:"name,omitempty"`
// 	ViewUsers  []string `json:"view_users,omitempty"`
// 	ModUsers   []string `json:"mod_users,omitempty"`
// }
// type Event struct {
// 	Label          string   `json:"label,omitempty"`
// 	Name           string   `json:"name,omitempty"`
// 	Description    string   `json:"description,omitempty"`
// 	Timestamp      string   `json:"timestamp,omitempty"`
// 	CalendarLabels []string `json:"calendar_labels,omitempty"`
// }

type HealthCheck struct {
	Users     int       `json:"users,omitempty"`
	Calendars int       `json:"calendars,omitempty"`
	Events    int       `json:"events,omitempty"`
	Time      time.Time `json:"time,omitempty"`
}
