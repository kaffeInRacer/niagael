package domain

import (
	"encoding/json"
	"time"
)

type ProductVariant struct {
	Id            string          `json:"id"`
	ProductId     string          `json:"product_id"`
	Name          string          `json:"name"`
	Price         int64           `json:"price"`
	Stock         int64           `json:"stock"`
	StockReserved int64           `json:"stock_reserved"`
	Attributes    json.RawMessage `json:"attributes"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     *time.Time      `json:"updated_at"`
	DeletedAt     *time.Time      `json:"deleted_at"`
}
