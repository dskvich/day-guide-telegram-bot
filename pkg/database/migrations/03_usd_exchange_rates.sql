-- +migrate Up
CREATE TABLE usd_exchange_rates (
     id INTEGER PRIMARY KEY,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     rub FLOAT,
     try FLOAT
);