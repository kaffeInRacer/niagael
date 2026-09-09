package domain

import (
	"errors"
	"time"
)

var ErrInvalidPaymentAmount = errors.New("invalid payment amount")

const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusCancelled = "cancelled"
)

type Payment struct {
	Id          string     `json:"id"`
	OrderId     string     `json:"order_id"`
	Amount      int64      `json:"amount"`
	Method      string     `json:"method"`
	Status      string     `json:"status"`
	SnapToken   string     `json:"snap_token"`
	RedirectUrl string     `json:"redirect_url"`
	VaNumber    *string    `json:"va_number"`
	PaidAt      *time.Time `json:"paid_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func PaymentStatusFromCallback(transactionStatus, fraudStatus string) string {
	switch transactionStatus {
	case "settlement":
		return PaymentStatusPaid
	case "capture":
		if fraudStatus == "accept" || fraudStatus == "" {
			return PaymentStatusPaid
		}
	case "deny", "failure":
		return PaymentStatusFailed
	case "cancel", "expire":
		return PaymentStatusCancelled
	}
	return PaymentStatusPending
}

func PaymentTransitionAllowed(current, next string) bool {
	return current == PaymentStatusPending && next != PaymentStatusPending
}
