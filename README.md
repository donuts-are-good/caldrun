# caldrun

try it out

```
# signup and get an api token
curl --request POST \
  --url 'http://caldrun.com/users?=' \
  --header 'Content-Type: application/json' \
  --data '{
	"username": "change-this-username"
}'
```


## usage

This http json API allows you to create and manage users, calendars and events.

### POST /users
#### Description

This endpoint allows for the creation of new users.
#### Parameters

Body:

- `username`: A unique string that identifies the user.

#### Responses

If successful, this endpoint will return a JSON object representing the user:

- `username`: The username of the user.
- `label`: The unique label of the user.
- `token`: The unique authentication token for the user.

### GET /calendars
#### Description

This endpoint retrieves the calendars for a particular user.
#### Headers

- `User-Token`: The authentication token for the user.

#### Responses

If successful, this endpoint will return an array of calendars that belong to the user. Each calendar is represented by a JSON object:

- `label`: The unique label of the calendar.
- `owner_label`: The label of the user who owns the calendar.
- `name`: The name of the calendar.
- `view_users`: An array of user labels who can view the calendar.
- `mod_users`: An array of user labels who can modify the calendar.

### POST /calendars
#### Description

This endpoint creates a new calendar for a particular user.
#### Headers

- `User-Token`: The authentication token for the user.

#### Parameters

Body:

- `name`: The name of the new calendar.

#### Responses

If successful, this endpoint will return a JSON object representing the new calendar:

- `label`: The unique label of the calendar.
- `owner_label`: The label of the user who owns the calendar.
- `name`: The name of the calendar.
- `view_users`: An array of user labels who can view the calendar.
- `mod_users`: An array of user labels who can modify the calendar.

### GET /events
#### Description

This endpoint retrieves the events for a particular user.
#### Headers

- `User-Token`: The authentication token for the user.

#### Responses

If successful, this endpoint will return an array of events that belong to the user. Each event is represented by a JSON object:

- `label`: The unique label of the event.
- `owner_label`: The label of the user who owns the event.
- `name`: The name of the event.
- `description`: A description of the event.
- `timestamp`: The time of the event in the format "yyyy-mm-dd hh:mm:ss".
- `calendar_labels`: An array of calendar labels to which the event belongs.

### POST /events
#### Description

This endpoint creates a new event for a particular user.
#### Headers

- `User-Token`: The authentication token for the user.

#### Parameters

Body:

- `name`: The name of the new event.
- `description`: A description of the event.
- `timestamp`: The time of the event in the format "yyyy-mm-dd hh:mm:ss".
- `calendar_labels`: An array of calendar labels to which the event belongs.

#### Responses

If successful, this endpoint will return a JSON object representing the new event:

- `label`: The unique label of the event.
- `owner_label`: The label of the user who owns the event.
- `name`: The name of the event.
- `description`: A description of the event.
- `timestamp`: The time of the event in the format "yyyy-mm-dd hh:mm:ss".
- `calendar_labels`: An array of calendar labels to which the event belongs.
## license

MIT License 2023 donuts-are-good, for more info see license.md
