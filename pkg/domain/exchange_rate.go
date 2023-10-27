package domain

import "time"

type ExchangeRate struct {
	Timestamp time.Time
	Pair      CurrencyPair
	Rate      float64
}
