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
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("No .env file found, using system env")
	// }
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

	bot, err := tgbotapi.NewBotAPI("8705284890:AAEEydfrYpAcKx1B2bu_mKbjW4JGIKKsFsk")
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	repo := repo.CreateRepo(pool)
	service := service.CreateService(repo)
	updatesChan := bot.GetUpdatesChan(updateConfig)
	tgHandler := handlers.CreateTgBotHandler(service, bot)
	for update := range updatesChan {
		tgHandler.HandleUpdate(update)
	}
}
