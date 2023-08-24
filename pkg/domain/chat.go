package domain

import "time"

type Chat struct {
	ID           int64
	RegisteredBy string
	RegisteredAt time.Time
}
