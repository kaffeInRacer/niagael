package kafka

type OrderEventItem struct {
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id"`
	Quantity  int     `json:"quantity"`
}

type OutboxEvent struct {
	Type    string           `json:"type"`
	OrderID string           `json:"order_id"`
	UserID  string           `json:"user_id"`
	Items   []OrderEventItem `json:"items"`
}
