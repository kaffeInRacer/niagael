package kafka

// OutboxEvent is the event published to the order-events topic via the outbox.
type OutboxEvent struct {
	Type    string           `json:"type"`
	OrderID string           `json:"order_id"`
	UserID  string           `json:"user_id"`
	Items   []OrderEventItem `json:"items"`
}

type OrderEventItem struct {
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id"`
	Quantity  int     `json:"quantity"`
}
