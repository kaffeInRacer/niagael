package domain

import "time"

type PromoUsage struct {
	Id        string     `json:"id"`
	PromoId   string     `json:"promo_id"`
	UserId    string     `json:"user_id"`
	Quantity  int        `json:"quantity"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
