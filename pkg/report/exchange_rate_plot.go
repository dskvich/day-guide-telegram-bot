package report

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

type ExchangeRateBulkFetcher interface {
	FetchHistoryRate(context.Context, domain.CurrencyPair, int) ([]domain.ExchangeRate, error)
	FetchLatestRate(context.Context, domain.CurrencyPair) (*domain.ExchangeRate, error)
}

type ExchangeRatePlotFormatter interface {
	Format(weather domain.ExchangeRate) string
}

type exchangeRatePlot struct {
	fetcher   ExchangeRateBulkFetcher
	formatter ExchangeRateFormatter
}

func NewExchangeRatePlot(
	fetcher ExchangeRateBulkFetcher,
	formatter ExchangeRateFormatter,
) *exchangeRatePlot {
	return &exchangeRatePlot{
		fetcher:   fetcher,
		formatter: formatter,
	}
}

func (e *exchangeRatePlot) Generate(ctx context.Context, pair domain.CurrencyPair) ([]byte, string, error) {
	// Create image
	graph := chart.Chart{}

	rates, err := e.fetcher.FetchHistoryRate(ctx, pair, 30)
	if err != nil {
		return nil, "", fmt.Errorf("fetching latest exchange rate for pair %s: %v", pair, err)
	}

	var xValues []time.Time
	var yValues []float64

	for _, rate := range rates {
		xValues = append(xValues, rate.Timestamp)
		yValues = append(yValues, rate.Rate)
	}

	graph.Series = append(graph.Series, chart.TimeSeries{
		Name:    fmt.Sprintf("%s - %s", pair.Base, pair.Quote),
		XValues: xValues,
		YValues: yValues,
		Style: chart.Style{
			StrokeColor: drawing.ColorRed, // will supercede defaults
			FillColor:   drawing.ColorRed.WithAlpha(64),
		},
	})

	graph.Elements = []chart.Renderable{chart.Legend(&graph)}
	graph.Height = 700
	graph.Width = 1024

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)

	// Create caption
	var sb strings.Builder
	rate, err := e.fetcher.FetchLatestRate(ctx, pair)
	if err != nil {
		return nil, "", fmt.Errorf("fetching latest exchange rate for pair %s: %v", pair, err)
	}

	sb.WriteString(e.formatter.Format(*rate))
	sb.WriteString("\n")

	return buffer.Bytes(), sb.String(), err
}
