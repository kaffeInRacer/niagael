package midtrans

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"strings"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransClient struct {
	Snap      snap.Client
	Core      coreapi.Client
	serverKey string
}

type TransactionRequest struct {
	OrderId        string
	Amount         int64
	Items          []ItemDetail
	Customer       CustomerDetail
	ExpiryDuration int64
	ExpiryUnit     string
}

type ItemDetail struct {
	Id    string
	Name  string
	Price int64
	Qty   int
}

type CustomerDetail struct {
	FirstName string
	Email     string
	Phone     string
}

type TransactionResponse struct {
	Token       string
	RedirectUrl string
	ExpireTime  string
}

func New(serverKey, clientKey, environment string) *MidtransClient {
	env := midtrans.Sandbox
	if environment == "production" {
		env = midtrans.Production
	}

	snapClient := snap.Client{}
	snapClient.New(serverKey, env)

	coreClient := coreapi.Client{}
	coreClient.New(serverKey, env)

	return &MidtransClient{
		Snap:      snapClient,
		Core:      coreClient,
		serverKey: serverKey,
	}
}

func (m *MidtransClient) CreateTransaction(req TransactionRequest) (*TransactionResponse, error) {
	var itemDetails []map[string]interface{}
	for _, item := range req.Items {
		itemDetails = append(itemDetails, map[string]interface{}{
			"id":       item.Id,
			"name":     item.Name,
			"price":    item.Price,
			"quantity": item.Qty,
		})
	}

	payload := snap.RequestParamWithMap{
		"transaction_details": map[string]interface{}{
			"order_id":     req.OrderId,
			"gross_amount": req.Amount,
		},
		"expiry": map[string]interface{}{
			"unit":     "hour",
			"duration": 1,
		},
	}

	if req.Customer.FirstName != "" || req.Customer.Email != "" || req.Customer.Phone != "" {
		customer := map[string]interface{}{
			"first_name": req.Customer.FirstName,
		}
		if req.Customer.Email != "" {
			customer["email"] = req.Customer.Email
		}
		if req.Customer.Phone != "" {
			customer["phone"] = req.Customer.Phone
		}
		payload["customer_details"] = customer
	}

	if len(itemDetails) > 0 {
		payload["item_details"] = itemDetails
	}

	if req.ExpiryDuration > 0 && req.ExpiryUnit != "" {
		payload["expiry"] = map[string]interface{}{
			"unit":     req.ExpiryUnit,
			"duration": req.ExpiryDuration,
		}
	}

	resp, err := m.Snap.CreateTransactionWithMap(&payload)
	if err != nil {
		return nil, err
	}

	token, _ := resp["token"].(string)
	redirectURL, _ := resp["redirect_url"].(string)

	return &TransactionResponse{
		Token:       token,
		RedirectUrl: redirectURL,
	}, nil
}

func (m *MidtransClient) GetTransactionStatus(orderId string) (*coreapi.TransactionStatusResponse, error) {
	resp, err := m.Core.CheckTransaction(orderId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *MidtransClient) VerifySignature(orderId, statusCode, grossAmount, signatureKey string) bool {
	sum := sha512.Sum512([]byte(orderId + statusCode + grossAmount + m.serverKey))
	expected := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(expected), []byte(strings.ToLower(signatureKey))) == 1
}
