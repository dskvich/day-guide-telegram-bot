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
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/gpt"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openexchangerates"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openweathermap"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/report"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/broadcaster"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/loader"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/telegram"
	telegrambot "github.com/sushkevichd/day-guide-telegram-bot/pkg/telegram"
)

type Config struct {
	TelegramBotToken          string  `env:"TELEGRAM_BOT_TOKEN,required"`
	OpenWeatherMapAPIKey      string  `env:"OPEN_WEATHER_MAP_API_KEY,required"`
	OpenExchangeRatesAPPID    string  `env:"OPEN_EXCHANGE_RATES_APP_ID,required"`
	ChatGPTTelegramBotURL     string  `env:"CHAT_GPT_TELEGRAM_BOT_URL" envDefault:"http://chatgpt-telegram-bot:8080"`
	TelegramAuthorizedUserIDs []int64 `env:"TELEGRAM_AUTHORIZED_USER_IDS" envSeparator:" "`
	QuotesRestAPIKey          string  `env:"QUOTES_REST_API_KEY,required"`
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

	db, err := database.NewSQLite()
	if err != nil {
		return nil, fmt.Errorf("creating db: %v", err)
	}

	bot, err := telegrambot.NewBot(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %v", err)
	}
	authenticator := auth.NewAuthenticator(cfg.TelegramAuthorizedUserIDs)

	gptClient, err := gpt.NewClient(cfg.ChatGPTTelegramBotURL)
	if err != nil {
		return nil, fmt.Errorf("creating gpt client: %v", err)
	}

	locations := []domain.Location{
		domain.SaintPetersburg,
		domain.Pitkyaranta,
		domain.Antalya,
	}
	weatherRepo := repository.NewWeatherRepository(db)
	weatherReportGenerator := report.NewWeather(locations, weatherRepo, &formatter.Weather{}, gptClient)

	pairs := []domain.CurrencyPair{
		{domain.USD, domain.RUB},
		{domain.USD, domain.TRY},
	}
	exchangeRateRepo := repository.NewExchangeRateRepository(db)
	exchangeRateReportGenerator := report.NewExchangeRate(pairs, exchangeRateRepo, &formatter.ExchangeRate{}, gptClient)

	moonPhaseRepo := repository.NewMoonPhaseRepository(db)
	moonPhaseReportGenerator := report.NewMoonPhase(moonPhaseRepo, &formatter.MoonPhase{}, gptClient)

	quoteRepo := repository.NewQuoteRepository(db)
	quoteReportGenerator := report.NewQuote(quoteRepo, &formatter.Quote{}, gptClient)

	chatRepository := repository.NewChatRepository(db)
	defaultHandler := handler.NewRegister(chatRepository)
	handlers := []command.Handler{
		handler.NewRegister(chatRepository),
		handler.NewWeather(weatherReportGenerator),
		handler.NewExchangeRate(exchangeRateReportGenerator),
		handler.NewMoonPhase(moonPhaseReportGenerator),
		handler.NewQuote(quoteReportGenerator),
	}

	dispatcher := command.NewDispatcher(handlers, defaultHandler)
	messagesCh := make(chan domain.Message)

	if svc, err = telegram.NewService(bot, authenticator, dispatcher, messagesCh); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	openWeatherClient := openweathermap.NewClient(cfg.OpenWeatherMapAPIKey)

	if svc, err = loader.NewService[*domain.Weather, domain.Location](
		"weather loader",
		locations,
		openWeatherClient,
		weatherRepo,
		30*time.Minute,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = broadcaster.NewService(
		"weather broadcaster",
		"0 5,10,15 * * *",
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
		pairs,
		openExchangeRatesClient,
		exchangeRateRepo,
		time.Hour,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = broadcaster.NewService(
		"exchange rate broadcaster",
		"10 5,10,15 * * *",
		chatRepository,
		exchangeRateReportGenerator,
		messagesCh,
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
		30*time.Minute,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = broadcaster.NewService(
		"moon phase broadcaster",
		"15 16 * * *",
		chatRepository,
		moonPhaseReportGenerator,
		messagesCh,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}
	/*
		quotesRestClient := quotesrest.NewClient(cfg.QuotesRestAPIKey)

		if svc, err = loader.NewService[*domain.Quote, struct{}](
			"quote loader",
			nil,
			quotesRestClient,
			quoteRepo,
			8*time.Hour,
		); err == nil {
			svcGroup = append(svcGroup, svc)
		} else {
			return nil, err
		}
	*/
	if svc, err = broadcaster.NewService(
		"quote broadcaster",
		"5 5 * * *",
		chatRepository,
		quoteReportGenerator,
		messagesCh,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	return svcGroup, nil
}
