-- +migrate Up
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    base TEXT,
    quote TEXT,
    rate FLOAT
);