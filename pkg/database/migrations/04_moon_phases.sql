-- +migrate Up
CREATE TABLE moon_phases (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    age INTEGER,
    names TEXT,
    phase TEXT,
    distance_to_earth FLOAT,
    illumination_prc INTEGER,
    distance_to_sun FLOAT
);