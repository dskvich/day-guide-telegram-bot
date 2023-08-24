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
	return &weatherRepository{
		db: db,
	}
}

func (repo *weatherRepository) CreateNew(ctx context.Context, w *domain.Weather) error {
	set := []string{"location", "temp", "temp_feel", "pressure", "humidity", "weather", "weather_verbose", "wind_speed", "wind_direction"}
	args := []any{w.Location, w.Temp, w.TempFeel, w.Pressure, w.Humidity, w.Weather, w.WeatherVerbose, w.WindSpeed, w.WindDirection}
	placeholder := []string{"?", "?", "?", "?", "?", "?", "?", "?", "?"}

	q := `insert into weather (` + strings.Join(set, ", ") + `) values (` + strings.Join(placeholder, ",") + `)`

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("creating a new weather: %v", err)
	}

	return nil
}
