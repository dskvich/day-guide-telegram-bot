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
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/googleai"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/workers"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/auth"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/database"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/farmsense"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/formatter"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openexchangerates"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openweathermap"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/report"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/telegram"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/telegram/command"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/workers/loader"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/workers/plotbroadcaster"
	telegramservice "github.com/sushkevichd/day-guide-telegram-bot/pkg/workers/telegram"
)

// Cron for daily messages - check cron on https://crontab.guru
const (
	weatherDailyCron      = "1 6 * * *"    // At 09:01 UTC+3
	exchangeRateDailyCron = "0 6,15 * * *" // At 9:00 and 18:00 UTC+3
	moonPhaseDailyCron    = "30 17 * * *"  // At 20:30 UTC+3
	holidayDailyCron      = "2 6 * * *"    // At 9:02 UTC+3
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
	domain.Antalya,
	domain.NhaTrang,
}

// Currency pairs for exchange rate calculations
var exchangeRatePairs = []domain.CurrencyPair{
	{domain.USD, domain.RUB},
}

type Config struct {
	TelegramBotToken          string  `env:"TELEGRAM_BOT_TOKEN,required"`
	OpenAIToken               string  `env:"OPEN_AI_TOKEN,required"`
	OpenWeatherMapAPIKey      string  `env:"OPEN_WEATHER_MAP_API_KEY,required"`
	OpenExchangeRatesAPPID    string  `env:"OPEN_EXCHANGE_RATES_APP_ID,required"`
	GoogleAIAPIKey            string  `env:"GOOGLE_AI_API_KEY,required"`
	TelegramAuthorizedUserIDs []int64 `env:"TELEGRAM_AUTHORIZED_USER_IDS" envSeparator:" "`
	PgURL                     string  `env:"DATABASE_URL"`
	PgHost                    string  `env:"DB_HOST" envDefault:"localhost:65433"`
	Port                      string  `env:"PORT" envDefault:"8080"`
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
	workerGroup, err := setupWorkers()
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

	return workerGroup.Start(ctx)
}

func setupWorkers() (workers.Group, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parsing env config: %v", err)
	}

	var worker workers.Worker
	var workerGroup workers.Group

	db, err := database.NewPostgres(cfg.PgURL, cfg.PgHost)
	if err != nil {
		return nil, fmt.Errorf("creating db: %v", err)
	}

	telegramClient, err := telegram.NewClient(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %v", err)
	}
	authenticator := auth.NewAuthenticator(cfg.TelegramAuthorizedUserIDs)

	/*openAIClient, err := openai.NewClient(cfg.OpenAIToken)
	if err != nil {
		return nil, fmt.Errorf("creating open AI client: %v", err)
	}*/

	googleAIClient, err := googleai.NewClient(cfg.GoogleAIAPIKey)
	if err != nil {
		return nil, fmt.Errorf("creating google AI client: %v", err)
	}

	weatherRepo := repository.NewWeatherRepository(db)
	weatherReportGenerator := report.NewWeather(weatherForecastLocations, weatherRepo, &formatter.Weather{})

	exchangeRateRepo := repository.NewExchangeRateRepository(db)
	exchangeRateFormatter := formatter.ExchangeRate{}
	exchangeRatePlotReportGenerator := report.NewExchangeRatePlot(exchangeRateRepo, &exchangeRateFormatter)

	moonPhaseRepo := repository.NewMoonPhaseRepository(db)
	moonPhaseReportGenerator := report.NewMoonPhase(moonPhaseRepo, &formatter.MoonPhase{})

	chatRepository := repository.NewChatRepository(db)

	holidayRepository := repository.NewHolidayRepository(db)
	holidayReportGenerator := report.NewHoliday(holidayRepository)

	hackerNewsService := service.NewsHackerNewsService()

	messagesCh := make(chan domain.Message)
	commands := []telegram.Command{
		command.NewGetHackerNews(hackerNewsService, googleAIClient, telegramClient),
		command.NewRegister(chatRepository, messagesCh),
		command.NewWeather(weatherReportGenerator, messagesCh),
		command.NewExchangeRate(exchangeRatePlotReportGenerator, exchangeRatePairs, messagesCh),
		command.NewMoonPhase(moonPhaseReportGenerator, messagesCh),
		command.NewHoliday(holidayReportGenerator, messagesCh),
	}

	commandDispatcher := telegram.NewCommandDispatcher(commands)

	if worker, err = telegramservice.NewService(telegramClient, authenticator, commandDispatcher, messagesCh); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	openWeatherClient := openweathermap.NewClient(cfg.OpenWeatherMapAPIKey)

	if worker, err = loader.NewService[*domain.Weather, domain.Location](
		"weather loader",
		weatherForecastLocations,
		openWeatherClient,
		weatherRepo,
		weatherPoolInterval,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	if worker, err = workers.NewBroadcaster(
		"weather broadcaster",
		weatherDailyCron,
		chatRepository,
		weatherReportGenerator,
		messagesCh,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	openExchangeRatesClient := openexchangerates.NewClient(cfg.OpenExchangeRatesAPPID)

	if worker, err = loader.NewService[*domain.ExchangeRate, domain.CurrencyPair](
		"exchange rate loader",
		exchangeRatePairs,
		openExchangeRatesClient,
		exchangeRateRepo,
		exchangeRatePoolInterval,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	if worker, err = plotbroadcaster.NewService(
		"exchange rate broadcaster",
		exchangeRateDailyCron,
		chatRepository,
		exchangeRatePlotReportGenerator,
		messagesCh,
		exchangeRatePairs,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	farmSenseClient := farmsense.NewClient()

	if worker, err = loader.NewService[*domain.MoonPhase, struct{}](
		"moon phase loader",
		nil,
		farmSenseClient,
		moonPhaseRepo,
		moonPhasePoolInterval,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	if worker, err = workers.NewBroadcaster(
		"moon phase broadcaster",
		moonPhaseDailyCron,
		chatRepository,
		moonPhaseReportGenerator,
		messagesCh,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	if worker, err = workers.NewBroadcaster(
		"holiday broadcaster",
		holidayDailyCron,
		chatRepository,
		holidayReportGenerator,
		messagesCh,
	); err == nil {
		workerGroup = append(workerGroup, worker)
	} else {
		return nil, err
	}

	return workerGroup, nil
}
