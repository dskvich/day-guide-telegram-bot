package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type moonPhaseRepository struct {
	db *sql.DB
}

func NewMoonPhaseRepository(db *sql.DB) *moonPhaseRepository {
	return &moonPhaseRepository{db: db}
}

func (repo *moonPhaseRepository) Save(ctx context.Context, m *domain.MoonPhase) error {
	columns := []string{"age", "names", "phase", "distance_to_earth", "illumination_prc", "distance_to_sun"}
	args := []any{m.Age, strings.Join(m.Names, ","), m.Phase, m.DistanceToEarth, m.IlluminationPrc, m.DistanceToSun}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	q := fmt.Sprintf("INSERT INTO moon_phases (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if _, err := repo.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	return nil
}

func (repo *moonPhaseRepository) FetchLatestPhase(ctx context.Context) (*domain.MoonPhase, error) {
	q := `
		select 
		    age,
			names,
			phase,
			distance_to_earth,
			illumination_prc,
			distance_to_sun
		from moon_phases
		order by created_at desc
		limit 1;
	`

	var m domain.MoonPhase
	var names string
	if err := repo.db.QueryRowContext(ctx, q).Scan(
		&m.Age,
		&names,
		&m.Phase,
		&m.DistanceToEarth,
		&m.IlluminationPrc,
		&m.DistanceToSun,
	); err != nil {
		return nil, fmt.Errorf("scanning row: %v", err)
	}

	m.Names = strings.Split(names, ",")

	return &m, nil
}
