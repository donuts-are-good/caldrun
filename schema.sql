CREATE TABLE users (
    label TEXT PRIMARY KEY,
    token TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL
);

CREATE TABLE calendars (
    label TEXT PRIMARY KEY,
    owner_label TEXT NOT NULL,
    name TEXT NOT NULL,
    view_users TEXT NOT NULL, -- Comma-separated user labels
    mod_users TEXT NOT NULL, -- Comma-separated user labels
    FOREIGN KEY (owner_label) REFERENCES users (label)
);

CREATE TABLE events (
    label TEXT PRIMARY KEY,
    owner_label TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    calendar_labels TEXT NOT NULL -- Comma-separated calendar labels
);
