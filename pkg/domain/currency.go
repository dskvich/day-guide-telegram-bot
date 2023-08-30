package domain

type Currency string

const (
	USD Currency = "USD"
	RUB Currency = "RUB"
	TRY Currency = "TRY"
)

func (c Currency) String() string {
	return string(c)
}

type CurrencyPair struct {
	Base  Currency
	Quote Currency
}
