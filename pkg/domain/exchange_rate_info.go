package domain

type ExchangeRateInfo struct {
	CurrentRate     *ExchangeRate
	PreviousDayRate *ExchangeRate
}

func (e *ExchangeRateInfo) PercentageChange() float64 {
	if e.PreviousDayRate.Rate == 0 {
		return 0
	}
	return ((e.CurrentRate.Rate - e.PreviousDayRate.Rate) / e.PreviousDayRate.Rate) * 100
}
