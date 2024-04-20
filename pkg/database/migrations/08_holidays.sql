-- +migrate Up
CREATE TABLE holidays (
     id SERIAL PRIMARY KEY,
     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
     date DATE NOT NULL,
     name TEXT NOT NULL
);

CREATE INDEX idx_holidays_date ON holidays (date);