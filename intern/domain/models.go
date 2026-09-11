package domain

import "time"

type User struct {
	Id           int64
	TelegramID   int64
	TelegramName string
	CreatedAt    time.Time
	Balance      float64
}
type VerificationCode struct {
	Id        int64
	UserId    int64
	Phone     string
	Code      string
	CreatedAt time.Time
	IsUsed    bool
}
type SimCard struct {
	Id        int64
	Phone     string
	CreatedAt time.Time
	IsActive  bool
	Message   string
}
type Rental struct {
	Id        int64     `json:"id"`
	CardId    int64     `json:"card_id"`
	UserId    int64     `json:"user_id"`
	Phone     string    `json:"phone"`
	RentedAt  time.Time `json:"rented_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Status    string    `json:"status"`
}
