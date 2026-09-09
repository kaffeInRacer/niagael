package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"

	grpcclient "kaffein/order-service/pkg/grpc/client"
	productpb "kaffein/order-service/proto/product"
)

type OrderEvent struct {
	Type    string           `json:"type"`
	OrderID string           `json:"order_id"`
	UserID  string           `json:"user_id"`
	Items   []OrderEventItem `json:"items"`
}

type OutboxRepo interface {
	ListUnprocessed(ctx context.Context, limit int) ([]OutboxEventRow, error)
	MarkProcessed(ctx context.Context, ids []int64) error
}

type OutboxEventRow struct {
	ID      int64
	Topic   string
	Key     string
	Payload []byte
}

type FulfillmentConsumer struct {
	reader        *kafka.Reader
	productClient *grpcclient.ProductClient
	pricingClient *grpcclient.DynamicPricingClient
}

func NewFulfillmentConsumer(brokers []string, topic, groupID string, productClient *grpcclient.ProductClient, pricingClient *grpcclient.DynamicPricingClient) *FulfillmentConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               brokers,
		Topic:                 topic,
		GroupID:               groupID,
		MinBytes:              1,
		MaxBytes:              10e6,
		WatchPartitionChanges: true,
	})
	return &FulfillmentConsumer{reader: reader, productClient: productClient, pricingClient: pricingClient}
}

func (c *FulfillmentConsumer) Run(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("kafka order-events read error: %v", err)
				continue
			}
			if err := c.handle(ctx, msg.Value); err != nil {
				log.Printf("kafka order-events handle error: %v payload=%s", err, string(msg.Value))
			}
		}
	}()
}

func (c *FulfillmentConsumer) handle(ctx context.Context, data []byte) error {
	var event OrderEvent
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("kafka order-events unmarshal error: %v", err)
		return nil
	}

	switch event.Type {
	case "payment.settled":
		items, err := c.toStockItems(event)
		if err != nil {
			return err
		}
		resp, err := c.productClient.ConfirmStock(ctx, event.OrderID, items)
		if err != nil {
			return err
		}
		if resp == nil || !resp.GetSuccess() {
			log.Printf("kafka: confirm stock failed for order %s: %s", event.OrderID, resp.GetMessage())
			return nil
		}
		log.Printf("kafka: confirmed stock for order %s", event.OrderID)
		return nil

	case "order.cancelled", "order.expired", "order.refunded":
		items, err := c.toStockItems(event)
		if err != nil {
			return err
		}
		if _, err := c.productClient.ReleaseStock(ctx, event.OrderID, items); err != nil {
			return err
		}
		if err := c.pricingClient.ReleasePricing(ctx, event.OrderID, event.UserID); err != nil {
			return err
		}
		log.Printf("kafka: released stock+pricing for %s order %s", event.Type, event.OrderID)
		return nil
	}
	return nil
}

func (c *FulfillmentConsumer) toStockItems(event OrderEvent) ([]*productpb.StockItem, error) {
	items := make([]*productpb.StockItem, 0, len(event.Items))
	for _, item := range event.Items {
		var variantID string
		if item.VariantID != nil {
			variantID = *item.VariantID
		}
		items = append(items, &productpb.StockItem{
			ProductId: item.ProductID,
			VariantId: variantID,
			Quantity:  int32(item.Quantity),
		})
	}
	return items, nil
}

func (c *FulfillmentConsumer) Close() error {
	return c.reader.Close()
}
