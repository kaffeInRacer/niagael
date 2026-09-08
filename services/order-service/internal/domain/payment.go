package domain

import "time"

type Payment struct {
	Id        string     `json:"id"`
	OrderId   string     `json:"order_id"`
	Amount    int64      `json:"amount"`
	Method    string     `json:"method"`
	Status    string     `json:"status"`
	SnapToken string     `json:"snap_token"`
	VaNumber  *string    `json:"va_number"`
	PaidAt    *time.Time `json:"paid_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
