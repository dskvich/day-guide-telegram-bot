-- +migrate Up
CREATE TABLE chats (
    id INTEGER PRIMARY KEY,
    registered_by TEXT,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);