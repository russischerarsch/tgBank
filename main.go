package main

import (
	dbconnection "bankTg/db_connection"
	"bankTg/intern/handlers"
	"bankTg/intern/metrics"
	"bankTg/intern/repo"
	"bankTg/intern/service"
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func filterUpdates(in tgbotapi.UpdatesChannel, svc *service.Service, log *slog.Logger) tgbotapi.UpdatesChannel {
	out := make(chan tgbotapi.Update)
	go func() {
		defer close(out)
		for u := range in {
			if shouldSkip(u, svc, log) {
				continue
			}
			out <- u
		}
	}()
	return out
}

func shouldSkip(u tgbotapi.Update, svc *service.Service, log *slog.Logger) bool {
	var userID string
	switch {
	case u.Message != nil:
		userID = strconv.FormatInt(u.Message.From.ID, 10)
	case u.CallbackQuery != nil:
		userID = strconv.FormatInt(u.CallbackQuery.From.ID, 10)
	default:
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	blocked, err := svc.IsBlocked(ctx, userID)
	if err != nil {
		log.Error("is_blocked", "err", err, "user", userID)
		return false
	}
	if blocked {
		log.Info("blocked user ignored", "user", userID)
	}
	return blocked
}

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	pool, err := dbconnection.CreateConnection(ctx)
	if err != nil {
		log.Fatal("DB:", err)
	}
	defer pool.Close()
	log.Println("DB connected")

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			stat := pool.Stat()
			metrics.DbPoolConnections.WithLabelValues("active").Set(float64(stat.AcquiredConns()))
			metrics.DbPoolConnections.WithLabelValues("idle").Set(float64(stat.IdleConns()))
		}
	}()

	key := os.Getenv("TG_TOKEN")
	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	repo := repo.CreateRepo(pool)
	svc := service.CreateService(repo)

	updatesChan := bot.GetUpdatesChan(updateConfig)
	updatesChan = filterUpdates(updatesChan, svc, logger)

	tgHandler := handlers.CreateTgBotHandler(svc, bot)
	for update := range updatesChan {
		tgHandler.HandleUpdate(update)
	}
}
