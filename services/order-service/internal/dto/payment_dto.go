package dto

type CreatePaymentDto struct {
	OrderId string
	Method  string `json:"method" validate:"required,def_enum=bank_transfer e_wallet qris va"`
}

type MidtransCallbackDto struct {
	OrderId           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionId     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	Bank              string `json:"bank"`
	VaNumber          string `json:"va_number"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	PaymentAmount     int64
}
