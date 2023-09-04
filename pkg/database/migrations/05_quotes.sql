-- +migrate Up
CREATE TABLE quotes (
     id INTEGER PRIMARY KEY,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     quote TEXT,
     author TEXT
);