package domain

import "time"

type Order struct {
	Id          string     `json:"id"`
	OrderRef    string     `json:"order_ref"`
	UserId      string     `json:"user_id"`
	AddressId   string     `json:"address_id"`
	TotalAmount int64      `json:"total_amount"`
	Status      string     `json:"status"`
	SnapToken   *string    `json:"snap_token"`
	ExpireTime  *time.Time `json:"expire_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`

	BuyerEmail string `json:"buyer_email,omitempty"`
}
