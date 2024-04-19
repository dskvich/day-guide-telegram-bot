package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v9"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/auth"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/command"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/command/handler"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/database"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/farmsense"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/formatter"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openai"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openexchangerates"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openweathermap"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/report"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/broadcaster"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/loader"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/plotbroadcaster"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/telegram"
	telegrambot "github.com/sushkevichd/day-guide-telegram-bot/pkg/telegram"
)

// Cron for daily messages - check cron on https://crontab.guru
const (
	weatherDailyCron      = "50 5 * * *"    // At 08:50 UTC+3
	exchangeRateDailyCron = "10 6,15 * * *" // At minute 10 past hour 9 and 18 UTC+3
	moonPhaseDailyCron    = "30 17 * * *"   // At 20:30 UTC+3
)

// Pool intervals for loaders
const (
	weatherPoolInterval      = 30 * time.Minute
	exchangeRatePoolInterval = 8 * time.Hour
	moonPhasePoolInterval    = 30 * time.Minute
)

// Locations for a weather forecast
var weatherForecastLocations = []domain.Location{
	domain.SaintPetersburg,
	domain.Pitkyaranta,
	domain.Antalya,
}

// Currency pairs for exchange rate calculations
var exchangeRatePairs = []domain.CurrencyPair{
	{domain.USD, domain.RUB},
	{domain.USD, domain.TRY},
}

type Config struct {
	TelegramBotToken          string  `env:"TELEGRAM_BOT_TOKEN,required"`
	OpenAIToken               string  `env:"OPEN_AI_TOKEN,required"`
	OpenWeatherMapAPIKey      string  `env:"OPEN_WEATHER_MAP_API_KEY,required"`
	OpenExchangeRatesAPPID    string  `env:"OPEN_EXCHANGE_RATES_APP_ID,required"`
	TelegramAuthorizedUserIDs []int64 `env:"TELEGRAM_AUTHORIZED_USER_IDS" envSeparator:" "`
	PgURL                     string  `env:"DATABASE_URL"`
	PgHost                    string  `env:"DB_HOST" envDefault:"localhost:65432"`
}

func main() {
	slog.SetDefault(logger.New(slog.LevelDebug))

	if err := runMain(); err != nil {
		slog.Error("shutting down due to error", logger.Err(err))
		return
	}
	slog.Info("shutdown complete")
}

func runMain() error {
	svcGroup, err := setupServices()
	if err != nil {
		return err
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP)
		select {
		case s := <-sigCh:
			slog.Info("shutting down due to signal", "signal", s.String())
			cancelFn()
		case <-ctx.Done():
		}
	}()

	return svcGroup.Run(ctx)
}

func setupServices() (service.Group, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parsing env config: %v", err)
	}

	var svc service.Service
	var svcGroup service.Group

	db, err := database.NewPostgres(cfg.PgURL, cfg.PgHost)
	if err != nil {
		return nil, fmt.Errorf("creating db: %v", err)
	}

	bot, err := telegrambot.NewBot(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %v", err)
	}
	authenticator := auth.NewAuthenticator(cfg.TelegramAuthorizedUserIDs)

	_, err = openai.NewClient(cfg.OpenAIToken)
	if err != nil {
		return nil, fmt.Errorf("creating open AI client: %v", err)
	}

	weatherRepo := repository.NewWeatherRepository(db)
	weatherReportGenerator := report.NewWeather(weatherForecastLocations, weatherRepo, &formatter.Weather{})

	exchangeRateRepo := repository.NewExchangeRateRepository(db)
	exchangeRateFormatter := formatter.ExchangeRate{}
	exchangeRatePlotReportGenerator := report.NewExchangeRatePlot(exchangeRateRepo, &exchangeRateFormatter)

	moonPhaseRepo := repository.NewMoonPhaseRepository(db)
	moonPhaseReportGenerator := report.NewMoonPhase(moonPhaseRepo, &formatter.MoonPhase{})

	chatRepository := repository.NewChatRepository(db)

	messagesCh := make(chan domain.Message)
	defaultHandler := handler.NewRegister(chatRepository, messagesCh)
	handlers := []command.Handler{
		handler.NewRegister(chatRepository, messagesCh),
		handler.NewWeather(weatherReportGenerator, messagesCh),
		handler.NewExchangeRate(exchangeRatePlotReportGenerator, exchangeRatePairs, messagesCh),
		handler.NewMoonPhase(moonPhaseReportGenerator, messagesCh),
	}

	dispatcher := command.NewDispatcher(handlers, defaultHandler)

	if svc, err = telegram.NewService(bot, authenticator, dispatcher, messagesCh); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	openWeatherClient := openweathermap.NewClient(cfg.OpenWeatherMapAPIKey)

	if svc, err = loader.NewService[*domain.Weather, domain.Location](
		"weather loader",
		weatherForecastLocations,
		openWeatherClient,
		weatherRepo,
		weatherPoolInterval,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = broadcaster.NewService(
		"weather broadcaster",
		weatherDailyCron,
		chatRepository,
		weatherReportGenerator,
		messagesCh,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	openExchangeRatesClient := openexchangerates.NewClient(cfg.OpenExchangeRatesAPPID)

	if svc, err = loader.NewService[*domain.ExchangeRate, domain.CurrencyPair](
		"exchange rate loader",
		exchangeRatePairs,
		openExchangeRatesClient,
		exchangeRateRepo,
		exchangeRatePoolInterval,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = plotbroadcaster.NewService(
		"exchange rate broadcaster",
		exchangeRateDailyCron,
		chatRepository,
		exchangeRatePlotReportGenerator,
		messagesCh,
		exchangeRatePairs,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	farmSenseClient := farmsense.NewClient()

	if svc, err = loader.NewService[*domain.MoonPhase, struct{}](
		"moon phase loader",
		nil,
		farmSenseClient,
		moonPhaseRepo,
		moonPhasePoolInterval,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = broadcaster.NewService(
		"moon phase broadcaster",
		moonPhaseDailyCron,
		chatRepository,
		moonPhaseReportGenerator,
		messagesCh,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	return svcGroup, nil
}
