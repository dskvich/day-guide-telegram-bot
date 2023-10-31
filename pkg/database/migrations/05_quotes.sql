-- +migrate Up
CREATE TABLE quotes (
     id SERIAL PRIMARY KEY,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     quote TEXT,
     author TEXT
);