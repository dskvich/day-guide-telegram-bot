package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v9"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/database"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/formatters"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/logger"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/openweathermap"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/repository"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/services"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/services/telegram"
	"github.com/sushkevichd/day-guide-telegram-bot/pkg/services/weather"
)

type Config struct {
	TelegramBotToken     string `env:"TELEGRAM_BOT_TOKEN,required"`
	OpenWeatherMapAPIKey string `env:"OPEN_WEATHER_MAP_API_KEY,required"`
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

func setupServices() (services.Group, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parsing env config: %v", err)
	}

	var svc services.Service
	var svcGroup services.Group

	db, err := database.NewSQLite()
	if err != nil {
		return nil, fmt.Errorf("creating db: %v", err)
	}

	chatRepository := repository.NewChatRepository(db)
	messagesCh := make(chan string)

	if svc, err = telegram.NewBotService(cfg.TelegramBotToken, chatRepository, messagesCh); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	openWeatherClient := openweathermap.NewClient(cfg.OpenWeatherMapAPIKey)
	weatherRepo := repository.NewWeatherRepository(db)
	locations := []domain.Location{
		domain.SaintPetersburg,
		domain.Pitkyaranta,
		domain.Antalya,
	}

	if svc, err = weather.NewLoaderService(openWeatherClient, weatherRepo, locations); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	if svc, err = weather.NewBroadcasterService(weatherRepo, locations, &formatters.TableFormatter{}, messagesCh); err == nil {
		svcGroup = append(svcGroup, svc)
	} else {
		return nil, err
	}

	return svcGroup, nil
}
