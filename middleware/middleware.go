package middleware

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Repository interface {
	IsBlocked(ctx context.Context, userID string) (bool, error)
}

func Middleware(repo Repository, log *slog.Logger) func(tgbotapi.Update) bool {
	return func(u tgbotapi.Update) bool {
		var userID string
		switch {
		case u.Message != nil:
			userID = u.Message.From.UserName
		case u.CallbackQuery != nil:
			userID = u.CallbackQuery.From.UserName
		default:
			return true
		}

		blocked, err := repo.IsBlocked(context.Background(), userID)
		if err != nil {
			log.Error("block middleware: is_blocked failed", "err", err, "user", userID)
			return true
		}
		if blocked {
			log.Info("blocked user ignored", "user", userID)
			return false
		}
		return true
	}

}
