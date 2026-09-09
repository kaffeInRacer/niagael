package domain

import "time"

type Address struct {
	Id         string     `json:"id"`
	UserId     string     `json:"user_id"`
	Street     string     `json:"street"`
	City       string     `json:"city"`
	Province   string     `json:"province"`
	PostalCode string     `json:"postal_code"`
	Country    string     `json:"country"`
	IsDefault  bool       `json:"is_default"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}
