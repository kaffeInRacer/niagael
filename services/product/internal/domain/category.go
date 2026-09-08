package domain

import "time"

type Category struct {
	Id           string     `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Description  string     `json:"description"`
	IsActive     bool       `json:"is_active"`
	ProductCount *int64     `json:"product_count,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}
