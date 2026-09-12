package service

import (
	"bankTg/intern/domain"
	"bankTg/intern/metrics"
	"bankTg/intern/repo"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Service struct {
	repo *repo.Repository
}

func CreateService(repo *repo.Repository) *Service {
	return &Service{repo: repo}
}
func (s *Service) CreateUser(ctx context.Context, tgID int64, tgName string) error {
	// 1. Проверяем, существует ли пользователь
	exists, err := s.repo.UserExistsByTelegramID(ctx, tgID)
	if err != nil {
		return fmt.Errorf("check user exists: %w", err)
	}

	// 2. Если существует — возвращаем специальную ошибку
	if exists {
		return domain.ErrUserAlreadyExists
	}

	// 3. Создаём нового пользователя
	user := &domain.User{
		TelegramID:   tgID,
		TelegramName: tgName,
		Balance:      0,
		CreatedAt:    time.Now().UTC(),
	}

	// 4. Сохраняем в БД
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *Service) RentCard(ctx context.Context, tgID int64, service string) (int, error) {
	user, err := s.repo.FindUserByTgID(ctx, tgID)
	if err != nil {
		slog.Warn("User not found for rent", "tgID", tgID, "error", err)
		return 0, fmt.Errorf("failed to find user, %w", err)

	}
	ids, err := s.repo.FindFreeCards(ctx, service)

	count := len(ids)
	if count == 0 {
		slog.Warn("Not enough free cards", "available", count)
		return count, fmt.Errorf("only %d cards available", count)
	}
	if err != nil {
		slog.Error("Failed to find free cards", "login", tgID, "error", err)
		return 0, fmt.Errorf("failed to find free cards, %w", err)
	}
	expirationTime := time.Now().Add(20 * time.Minute)

	if err := s.repo.RentCard(ctx, user, ids, expirationTime); err != nil {
		slog.Error("Failed to rent card", "login", tgID, "error", err)
		return 0, fmt.Errorf("failed to rent a card, %w", err)
	}
	slog.Info("Cards rented", "login", tgID, "count", count)
	metrics.RentalsCreatedTotal.Inc()
	metrics.RentalsActive.Inc()
	metrics.CardsFree.Sub(float64(len(ids)))
	metrics.CardsBusy.Add(float64(len(ids)))
	return count, nil
}
func (s *Service) StartExpirationChecker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			count, err := s.repo.ExpireRentals(ctx)
			if err != nil {
				slog.Error("Expire rentals failed", "error", err)
				continue
			}
			if count > 0 {
				metrics.RentalsActive.Sub(float64(count))
				metrics.CardsBusy.Sub(float64(count))
				metrics.CardsFree.Add(float64(count))
				slog.Info("Rentals expired", "count", count)
			}
		}
	}()
}
func (s *Service) GetExpiredCards(ctx context.Context, id int64) ([]*domain.Rental, error) {

	cards, err := s.repo.GetExpiredCards(ctx, id)
	if err != nil {
		slog.Error("Failed to get expired cards", "user_id", id, "error", err)
		return nil, fmt.Errorf("get expired cards: %w", err)
	}
	if len(cards) == 0 {
		slog.Warn("No expired cards for user", "user_id", id)
		return nil, domain.ErrNoExpiredCards
	}
	slog.Info("Got expired cards", "user_id", id, "count", len(cards))
	return cards, nil
}

func (s *Service) GetActiveCards(ctx context.Context, id int64) ([]*domain.Rental, error) {

	cards, err := s.repo.GetActiveCards(ctx, id)
	if err != nil {
		slog.Error("Failed to get active cards", "user_id", id, "error", err)
		return nil, fmt.Errorf("get active cards: %w", err)
	}
	if len(cards) == 0 {
		slog.Warn("No active cards for user", "user_id", id)
		return nil, domain.ErrNoActiveCards
	}
	slog.Info("Got active cards", "user_id", id, "count", len(cards))
	return cards, nil
}
func (s *Service) GetBalance(ctx context.Context, telegramID int64) (float64, error) {
	user, err := s.repo.GetBalance(ctx, telegramID)
	if err != nil {
		slog.Error("Failed to get balance", "user_id", user.Id, "error", err)
		return 0, fmt.Errorf("get balance error: %w", err)
	}
	return user.Balance, nil
}
func (s *Service) IsBlocked(ctx context.Context, userID string) (bool, error) {
	return s.repo.IsBlocked(ctx, userID)
}
