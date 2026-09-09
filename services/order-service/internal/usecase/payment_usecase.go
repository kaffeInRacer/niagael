package usecase

import (
	"context"

	"errors"
	"fmt"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/internal/interfaces/IUseCase"
	grpcclient "kaffein/order-service/pkg/grpc/client"
	"kaffein/order-service/pkg/kafka"
	"kaffein/order-service/pkg/midtrans"
	"kaffein/order-service/pkg/postgresql"
	productpb "kaffein/order-service/proto/product"
	"kaffein/order-service/utils/constants"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type OutboxTxRepo interface {
	InsertWithTx(ctx context.Context, tx postgresql.DBTX, topic, key string, payload any) error
}

type paymentUseCase struct {
	paymentRepo   IRepository.PaymentRepository
	orderRepo     IRepository.OrderRepository
	orderItemRepo IRepository.OrderItemRepository
	midtrans      paymentGateway
	productClient stockClient
	pricingClient pricingReleaseClient
	outbox        OutboxTxRepo
}

type stockClient interface {
	ConfirmStock(context.Context, string, []*productpb.StockItem) (*productpb.ConfirmStockResponse, error)
	ReleaseStock(context.Context, string, []*productpb.StockItem) (*productpb.ReleaseStockResponse, error)
}

type paymentGateway interface {
	CreateTransaction(midtrans.TransactionRequest) (*midtrans.TransactionResponse, error)
	RedirectUrl(string) string
	VerifySignature(string, string, string, string) bool
}

type pricingReleaseClient interface {
	ReleasePricing(context.Context, string, string) error
}

func NewPaymentUseCase(paymentRepo IRepository.PaymentRepository, orderRepo IRepository.OrderRepository, orderItemRepo IRepository.OrderItemRepository, midtrans *midtrans.MidtransClient, productClient *grpcclient.ProductClient, pricingClient *grpcclient.DynamicPricingClient, outbox OutboxTxRepo) IUseCase.PaymentUseCase {
	return &paymentUseCase{
		paymentRepo:   paymentRepo,
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		midtrans:      midtrans,
		productClient: productClient,
		pricingClient: pricingClient,
		outbox:        outbox,
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

	payment := domain.Payment{
		Id:      uuid.New().String(),
		OrderId: args.OrderId,
		Amount:  order.TotalAmount,
		Method:  args.Method,
		Status:  domain.PaymentStatusPending,
	}
	localPayment, err := uc.paymentRepo.GetOrCreate(ctx, payment)
	if err != nil {
		return "", err
	}
	if localPayment == nil {
		return "", errors.New(constants.ErrPaymentNotFound)
	}
	if localPayment.Amount != order.TotalAmount {
		return "", domain.ErrInvalidPaymentAmount
	}
	if localPayment.Status != domain.PaymentStatusPending {
		return "", errors.New(constants.ErrOrderAlreadyPaid)
	}

	return uc.paymentRepo.InitiateWithTx(ctx, order.Id, func(payment *domain.Payment) (string, string, error) {
		if payment.SnapToken != "" {
			return payment.SnapToken, uc.midtrans.RedirectUrl(payment.SnapToken), nil
		}
		orderItems, err := uc.orderItemRepo.ReadByOrderId(ctx, order.Id)
		if err != nil {
			return "", "", err
		}
		items := make([]midtrans.ItemDetail, 0, len(orderItems))
		for _, item := range orderItems {
			items = append(items, midtrans.ItemDetail{Id: item.ProductId, Name: item.ProductName, Price: item.FinalPrice / int64(item.Quantity), Qty: item.Quantity})
		}
		resp, err := uc.midtrans.CreateTransaction(midtrans.TransactionRequest{
			OrderId: order.Id, Amount: order.TotalAmount, Items: items,
			ExpiryDuration: 1, ExpiryUnit: "hour", Customer: midtrans.CustomerDetail{FirstName: "Customer"},
		})
		if err != nil {
			return "", "", err
		}
		return resp.Token, resp.RedirectUrl, nil
	})
}

func (uc *paymentUseCase) Callback(ctx context.Context, args dto.MidtransCallbackDto) error {
	if !uc.midtrans.VerifySignature(args.OrderId, args.StatusCode, args.GrossAmount, args.SignatureKey) {
		return errors.New(constants.ErrInvalidPaymentSignature)
	}

	amount, err := parseGrossAmount(args.GrossAmount)
	if err != nil {
		return domain.ErrInvalidPaymentAmount
	}
	newStatus := domain.PaymentStatusFromCallback(args.TransactionStatus, args.FraudStatus)

	_, err = uc.paymentRepo.ProcessCallbackWithTx(ctx, args.OrderId, amount, newStatus, func(tx postgresql.DBTX) error {
		orderItems, err := uc.orderItemRepo.ReadByOrderId(ctx, args.OrderId)
		if err != nil {
			return err
		}

		if newStatus == domain.PaymentStatusPaid {
			event := kafka.OutboxEvent{
				Type:    "payment.settled",
				OrderID: args.OrderId,
			}
			for _, item := range orderItems {
				event.Items = append(event.Items, kafka.OrderEventItem{
					ProductID: item.ProductId,
					VariantID: item.VariantId,
					Quantity:  item.Quantity,
				})
			}
			return uc.outbox.InsertWithTx(ctx, tx, "order-events", args.OrderId, event)
		}

		var orderUserID string
		order, err := uc.orderRepo.ReadById(ctx, args.OrderId)
		if err != nil {
			return err
		}
		if order == nil {
			return errors.New(constants.ErrOrderNotFound)
		}
		orderUserID = order.UserId

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

		resp, err := uc.productClient.ReleaseStock(ctx, args.OrderId, stockItems)
		if err != nil {
			return err
		}
		if resp == nil || !resp.GetSuccess() {
			if resp == nil {
				return errors.New("release stock returned empty response")
			}
			return fmt.Errorf("release stock failed: %s", resp.GetMessage())
		}
		if err := uc.pricingClient.ReleasePricing(ctx, args.OrderId, orderUserID); err != nil {
			return err
		}
		return nil
	})
	return err
}

func parseGrossAmount(value string) (int64, error) {
	integer, fraction, found := strings.Cut(value, ".")
	if integer == "" || strings.HasPrefix(integer, "+") || strings.HasPrefix(integer, "-") {
		return 0, domain.ErrInvalidPaymentAmount
	}
	if found && (fraction == "" || strings.Trim(fraction, "0") != "") {
		return 0, domain.ErrInvalidPaymentAmount
	}
	amount, err := strconv.ParseInt(integer, 10, 64)
	if err != nil {
		return 0, domain.ErrInvalidPaymentAmount
	}
	return amount, nil
}
