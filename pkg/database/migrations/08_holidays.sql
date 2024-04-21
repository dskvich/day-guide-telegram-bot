-- +migrate Up
CREATE TABLE holidays (
     id SERIAL PRIMARY KEY,
     created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
     order_number INT,
     date DATE NOT NULL,
     name TEXT NOT NULL
);

CREATE INDEX idx_holidays_order_number ON holidays (order_number);
CREATE INDEX idx_holidays_date ON holidays (date);