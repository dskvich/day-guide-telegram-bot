package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type weatherRepository struct {
	db *sql.DB
}

func NewWeatherRepository(db *sql.DB) *weatherRepository {
	return &weatherRepository{db: db}
}

func (repo *weatherRepository) Save(ctx context.Context, w *domain.Weather) error {
	set := []string{"location", "temp", "temp_feel", "pressure", "humidity", "weather", "weather_verbose", "wind_speed", "wind_direction"}
	args := []any{w.Location, w.Temp, w.TempFeel, w.Pressure, w.Humidity, w.Weather, w.WeatherVerbose, w.WindSpeed, w.WindDirection}
	placeholder := []string{"?", "?", "?", "?", "?", "?", "?", "?", "?"}

	q := `insert into weather (` + strings.Join(set, ", ") + `) values (` + strings.Join(placeholder, ",") + `)`

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	return nil
}

func (repo *weatherRepository) FetchLatestByLocation(ctx context.Context, location domain.Location) (*domain.Weather, error) {
	q := `
		select 
		    location,
		    temp,
		    temp_feel,
		    pressure,
			humidity,
			weather,
			weather_verbose,
			wind_speed,
			wind_direction
		from weather
		where location = ?
		order by created_at desc
		limit 1;
	`

	var w domain.Weather
	if err := repo.db.QueryRowContext(ctx, q, location).Scan(
		&w.Location,
		&w.Temp,
		&w.TempFeel,
		&w.Pressure,
		&w.Humidity,
		&w.Weather,
		&w.WeatherVerbose,
		&w.WindSpeed,
		&w.WindDirection,
	); err != nil {
		return nil, fmt.Errorf("scanning row: %v", err)
	}

	return &w, nil
}
