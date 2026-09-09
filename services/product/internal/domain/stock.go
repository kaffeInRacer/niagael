package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"

	"kaffein/product-service/utils/constants"
)

type StockOperation string

const (
	StockReserve StockOperation = "reserve"
	StockConfirm StockOperation = "confirm"
	StockRelease StockOperation = "release"
)

const (
	stockStateReserved  = "reserved"
	stockStateConfirmed = "confirmed"
	stockStateReleased  = "released"
)

type StockItem struct {
	ProductID string
	VariantID string
	Quantity  int32
}

type CanonicalStockItem struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int64  `json:"quantity"`
}

func CanonicalizeStockItems(rawItems []StockItem) ([]CanonicalStockItem, string, []byte, error) {
	if len(rawItems) == 0 {
		return nil, "", nil, errors.New(constants.ErrInvalidStockRequest)
	}

	aggregated := make(map[string]CanonicalStockItem, len(rawItems))
	for _, item := range rawItems {
		if item.ProductID == "" || item.Quantity <= 0 {
			return nil, "", nil, errors.New(constants.ErrInvalidStockRequest)
		}

		key := item.ProductID + "\x00" + item.VariantID
		entry := aggregated[key]
		entry.ProductID, entry.VariantID = item.ProductID, item.VariantID
		entry.Quantity += int64(item.Quantity)
		aggregated[key] = entry
	}

	items := make([]CanonicalStockItem, 0, len(aggregated))
	for _, item := range aggregated {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ProductID == items[j].ProductID {
			return items[i].VariantID < items[j].VariantID
		}
		return items[i].ProductID < items[j].ProductID
	})

	payload, err := json.Marshal(items)
	if err != nil {
		return nil, "", nil, err
	}
	sum := sha256.Sum256(payload)

	return items, hex.EncodeToString(sum[:]), payload, nil
}

func InitialStockState(operation StockOperation) string {
	if operation == StockRelease {
		return stockStateReleased
	}
	return stockStateReserved
}

func StockTransition(operation StockOperation, isNew bool, state string) (skip bool, nextStatus string, err error) {
	switch operation {
	case StockReserve:
		if !isNew {
			return true, "", nil
		}
		return false, "", nil

	case StockConfirm:
		if isNew {
			return false, "", errors.New(constants.ErrStockStateConflict + ": reservation does not exist")
		}
		if state == stockStateConfirmed {
			return true, "", nil
		}
		if state != stockStateReserved {
			return false, "", errors.New(constants.ErrStockStateConflict + ": cannot confirm " + state + " reservation")
		}
		return false, stockStateConfirmed, nil

	case StockRelease:
		if isNew || state == stockStateReleased {
			return true, "", nil
		}
		if state != stockStateReserved {
			return false, "", errors.New(constants.ErrStockStateConflict + ": cannot release " + state + " reservation")
		}
		return false, stockStateReleased, nil

	default:
		return false, "", errors.New(constants.ErrInvalidStockRequest)
	}
}
