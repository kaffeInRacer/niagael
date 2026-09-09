package kafka

import (
	"context"
	"log"

	grpcclient "kaffein/order-service/pkg/grpc/client"
	productpb "kaffein/order-service/proto/product"
)

type FulfillmentConsumer struct {
	productClient *grpcclient.ProductClient
	pricingClient *grpcclient.DynamicPricingClient
}

func NewFulfillmentConsumer(productClient *grpcclient.ProductClient, pricingClient *grpcclient.DynamicPricingClient) *FulfillmentConsumer {
	return &FulfillmentConsumer{productClient: productClient, pricingClient: pricingClient}
}

// Handle processes one order-events message. It is wired to event.Consume.
func (c *FulfillmentConsumer) Handle(ctx context.Context, data []byte) error {
	event, err := DecodeOutboxEvent(data)
	if err != nil {
		log.Printf("kafka order-events unmarshal error: %v", err)
		return nil
	}

	switch event.Type {
	case "payment.settled":
		items := ConvertToStockItems(event.Items)
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
		items := ConvertToStockItems(event.Items)
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

func ConvertToStockItems(eventItems []OrderEventItem) []*productpb.StockItem {
	items := make([]*productpb.StockItem, 0, len(eventItems))
	for _, item := range eventItems {
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
	return items
}
