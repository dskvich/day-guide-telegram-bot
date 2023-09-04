package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
)

type chatRepository struct {
	db *sql.DB
}

var ErrChatAlreadyExists = errors.New("chat with the given ID already exists")

func NewChatRepository(db *sql.DB) *chatRepository {
	return &chatRepository{db: db}
}

func (repo *chatRepository) Save(ctx context.Context, chat *domain.Chat) error {
	q := `insert into chats(id, registered_by) values(?, ?)`

	if _, err := repo.db.ExecContext(ctx, q, chat.ID, chat.RegisteredBy); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrChatAlreadyExists
		}
		return fmt.Errorf("creating a new chat: %v", err)
	}

	return nil
}

func (repo *chatRepository) GetIDs(ctx context.Context) ([]int64, error) {
	q := `select id from chats`

	rows, err := repo.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("querying chat IDs: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Warn("Failed to close rows", logger.Err(err))
		}
	}()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning rows: %v", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}
