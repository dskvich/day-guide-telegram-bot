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

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/command"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/command/handler"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/database"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/formatter"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/gpt"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openexchangerates"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openweathermap"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/report"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/broadcaster"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/exchangerates"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/telegram"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/service/weather"
	telegrambot "github.com/sushkevichd/day-guide-telegram-bot/pkg/telegram"
)

type Config struct {
	TelegramBotToken       string `env:"TELEGRAM_BOT_TOKEN,required"`
	OpenWeatherMapAPIKey   string `env:"OPEN_WEATHER_MAP_API_KEY,required"`
	OpenExchangeRatesAPPID string `env:"OPEN_EXCHANGE_RATES_APP_ID,required"`
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

	chatRepository := repository.NewChatRepository(db)
	defaultHandler := handler.NewRegister(chatRepository)
	handlers := []command.Handler{
		handler.NewRegister(chatRepository),
	}

	dispatcher := command.NewDispatcher(handlers, defaultHandler)
	messagesCh := make(chan domain.Message)

	if svc, err = telegram.NewService(bot, dispatcher, messagesCh); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	openWeatherClient := openweathermap.NewClient(cfg.OpenWeatherMapAPIKey)
	weatherRepo := repository.NewWeatherRepository(db)
	locations := []domain.Location{
		domain.SaintPetersburg,
		domain.Pitkyaranta,
		domain.Phuket,
		domain.Antalya,
	}

	if svc, err = weather.NewLoaderService(
		openWeatherClient,
		weatherRepo,
		locations,
		30*time.Minute,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	gptClient := gpt.NewClient()
	weatherReportGenerator := report.NewWeather(locations, weatherRepo, &formatter.Weather{}, gptClient)

	if svc, err = broadcaster.NewBroadcasterService(
		"weather",
		"0 6,10,15 * * *",
		chatRepository,
		weatherReportGenerator,
		messagesCh,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	openExchangeRatesClient := openexchangerates.NewClient(cfg.OpenExchangeRatesAPPID)
	exchangeRatesRepo := repository.NewExchangeRatesRepository(db)

	if svc, err = exchangerates.NewLoaderService(
		openExchangeRatesClient,
		exchangeRatesRepo,
		time.Hour,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	exchangeRatesReportGenerator := report.NewExchangeRates(exchangeRatesRepo, &formatter.ExchageRates{})

	if svc, err = broadcaster.NewBroadcasterService(
		"exchange rates",
		"0 6,8,10,12,14,16 * * *",
		chatRepository,
		exchangeRatesReportGenerator,
		messagesCh,
	); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	return svcGroup, nil
}
