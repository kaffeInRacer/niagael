package usecase

import (
	"context"
	"errors"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/internal/interfaces/IUseCase"
	grpcclient "kaffein/order-service/pkg/grpc/client"
	"kaffein/order-service/pkg/midtrans"
	"kaffein/order-service/proto/product"
	"kaffein/order-service/utils/constants"

	"github.com/google/uuid"
)

type paymentUseCase struct {
	paymentRepo   IRepository.PaymentRepository
	orderRepo     IRepository.OrderRepository
	orderItemRepo IRepository.OrderItemRepository
	midtrans      *midtrans.MidtransClient
	productClient *grpcclient.ProductClient
}

func NewPaymentUseCase(paymentRepo IRepository.PaymentRepository, orderRepo IRepository.OrderRepository, orderItemRepo IRepository.OrderItemRepository, midtrans *midtrans.MidtransClient, productClient *grpcclient.ProductClient) IUseCase.PaymentUseCase {
	return &paymentUseCase{
		paymentRepo:   paymentRepo,
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		midtrans:      midtrans,
		productClient: productClient,
	}
}

func (uc *paymentUseCase) Create(ctx context.Context, args dto.CreatePaymentDto) (string, error) {
	order, err := uc.orderRepo.ReadById(ctx, args.OrderId)
	if err != nil {
		return "", err
	}

	if order == nil {
		return "", errors.New(constants.ErrOrderNotFound)
	}

	if order.Status != "pending" {
		return "", errors.New(constants.ErrOrderAlreadyPaid)
	}

	orderItems, err := uc.orderItemRepo.ReadByOrderId(ctx, order.Id)
	if err != nil {
		return "", err
	}

	payment := domain.Payment{
		Id:      uuid.New().String(),
		OrderId: args.OrderId,
		Amount:  order.TotalAmount,
		Method:  args.Method,
		Status:  "pending",
	}

	var items []midtrans.ItemDetail
	for _, item := range orderItems {
		unitPrice := item.FinalPrice / int64(item.Quantity)
		items = append(items, midtrans.ItemDetail{
			Id:    item.ProductId,
			Name:  item.ProductName,
			Price: unitPrice,
			Qty:   item.Quantity,
		})
	}

	// Snap: blanket expiry 1 hour untuk semua payment method
	// Payment method expiry (VA 24 jam, GoPay 15 menit) diatur di Midtrans Dashboard
	req := midtrans.TransactionRequest{
		OrderId:        order.Id,
		Amount:         order.TotalAmount,
		Items:          items,
		ExpiryDuration: 1,
		ExpiryUnit:     "hour",
		Customer: midtrans.CustomerDetail{
			FirstName: "Customer",
		},
	}

	resp, err := uc.midtrans.CreateTransaction(req)
	if err != nil {
		return "", err
	}

	payment.SnapToken = resp.Token

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return "", err
	}

	// Update order with snap_token
	order.SnapToken = resp.Token
	if err := uc.orderRepo.Update(ctx, order.Id, *order); err != nil {
		return "", err
	}

	return resp.RedirectUrl, nil
}

func (uc *paymentUseCase) Callback(ctx context.Context, args dto.MidtransCallbackDto) error {
	if !uc.midtrans.VerifySignature(args.OrderId, args.StatusCode, args.GrossAmount, args.SignatureKey) {
		return errors.New(constants.ErrInvalidPaymentSignature)
	}

	payment, err := uc.paymentRepo.ReadByOrderId(ctx, args.OrderId)
	if err != nil {
		return err
	}

	if payment == nil {
		return errors.New(constants.ErrPaymentNotFound)
	}

	// Update payment status based on Midtrans status
	newStatus := "pending"
	switch args.TransactionStatus {
	case "settlement":
		newStatus = "paid"
	case "capture":
		if args.FraudStatus == "accept" || args.FraudStatus == "" {
			newStatus = "paid"
		}
	case "deny", "failure":
		newStatus = "failed"
	case "cancel", "expire":
		newStatus = "cancelled"
	}

	if err := uc.paymentRepo.UpdateStatus(ctx, payment.Id, newStatus); err != nil {
		return err
	}

	// Update order status
	orderStatus := "pending"
	switch newStatus {
	case "paid":
		orderStatus = "paid"
	case "failed", "cancelled":
		orderStatus = "cancelled"
	}

	order, err := uc.orderRepo.ReadById(ctx, args.OrderId)
	if err != nil {
		return err
	}

	if order != nil {
		order.Status = orderStatus
		if err := uc.orderRepo.Update(ctx, order.Id, *order); err != nil {
			return err
		}

		// Get order items for stock operations
		orderItems, err := uc.orderItemRepo.ReadByOrderId(ctx, order.Id)
		if err != nil {
			return err
		}

		// Prepare stock items
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

		// Handle stock based on payment status
		if newStatus == "paid" {
			// Payment successful, confirm stock
			_, err := uc.productClient.ConfirmStock(ctx, order.Id, stockItems)
			if err != nil {
				return err
			}
		} else if newStatus == "failed" || newStatus == "cancelled" {
			// Payment failed or cancelled, release stock
			_, err := uc.productClient.ReleaseStock(ctx, order.Id, stockItems)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
