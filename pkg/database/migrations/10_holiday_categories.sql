-- +migrate Up
CREATE TABLE IF NOT EXISTS holiday_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);