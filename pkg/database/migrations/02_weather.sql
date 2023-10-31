-- +migrate Up
CREATE TABLE weather (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    location TEXT,
    temp FLOAT,
    temp_feel FLOAT,
    pressure INTEGER,
    humidity INTEGER,
    weather TEXT,
    weather_verbose TEXT,
    wind_speed FLOAT,
    wind_direction TEXT
);