package main

import (
	"context"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/repository"
	"kaffein/order-service/pkg/postgresql"
	"kaffein/order-service/proto/product"
	"time"
)

func (app *application) startOrderExpiryJob(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			app.logger.Info().Msg("stopping order expiry job")
			return
		case <-ticker.C:
			app.cancelExpiredOrders(ctx)
		}
	}
}

func (app *application) cancelExpiredOrders(ctx context.Context) {
	store := postgresql.NewStore(app.pgx)
	orderRepo := repository.NewOrderRepository(store)
	orderItemRepo := repository.NewOrderItemRepository(store)

	// Find orders that are pending and past their expire_time
	query := `
		SELECT id, user_id, address_id, total_amount, status, expire_time, created_at 
		FROM "orders" 
		WHERE status = 'pending' 
		AND expire_time IS NOT NULL 
		AND expire_time < NOW()
		AND deleted_at IS NULL
	`

	rows, err := app.pgx.Query(ctx, query)
	if err != nil {
		app.logger.Error().Err(err).Msg("failed to query expired orders")
		return
	}
	defer rows.Close()

	var expiredOrders []struct {
		Id          string
		UserId      string
		AddressId   string
		TotalAmount int64
		Status      string
		ExpireTime  *time.Time
		CreatedAt   time.Time
	}

	for rows.Next() {
		var order struct {
			Id          string
			UserId      string
			AddressId   string
			TotalAmount int64
			Status      string
			ExpireTime  *time.Time
			CreatedAt   time.Time
		}
		if err := rows.Scan(
			&order.Id,
			&order.UserId,
			&order.AddressId,
			&order.TotalAmount,
			&order.Status,
			&order.ExpireTime,
			&order.CreatedAt,
		); err != nil {
			app.logger.Error().Err(err).Msg("failed to scan expired order")
			continue
		}
		expiredOrders = append(expiredOrders, order)
	}

	if len(expiredOrders) == 0 {
		return
	}

	app.logger.Info().Int("count", len(expiredOrders)).Msg("cancelling expired orders")

	for _, order := range expiredOrders {
		// Update order status to cancelled
		orderDomain := domain.Order{
			Id:          order.Id,
			UserId:      order.UserId,
			AddressId:   order.AddressId,
			TotalAmount: order.TotalAmount,
			Status:      "cancelled",
			ExpireTime:  order.ExpireTime,
			CreatedAt:   order.CreatedAt,
		}
		if err := orderRepo.Update(ctx, order.Id, orderDomain); err != nil {
			app.logger.Error().Err(err).Str("order_id", order.Id).Msg("failed to update order status")
			continue
		}

		// Get order items for stock release
		orderItems, err := orderItemRepo.ReadByOrderId(ctx, order.Id)
		if err != nil {
			app.logger.Error().Err(err).Str("order_id", order.Id).Msg("failed to get order items")
			continue
		}

		// Prepare stock items for release
		var stockItems []*productpb.StockItem
		for _, item := range orderItems {
			var variantId string
			if item.VariantId != nil {
				variantId = *item.VariantId
			}
			stockItem := &productpb.StockItem{
				ProductId: item.ProductId,
				VariantId: variantId,
				Quantity:  int32(item.Quantity),
			}
			stockItems = append(stockItems, stockItem)
		}

		// Release stock
		if _, err := app.productClient.ReleaseStock(ctx, order.Id, stockItems); err != nil {
			app.logger.Error().Err(err).Str("order_id", order.Id).Msg("failed to release stock")
			continue
		}

		app.logger.Info().Str("order_id", order.Id).Msg("cancelled expired order and released stock")
	}
}
