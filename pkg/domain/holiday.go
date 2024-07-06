package domain

import "time"

type Holiday struct {
	OrderNumber int
	Name        string
	Date        time.Time
	Categories  []string
}
