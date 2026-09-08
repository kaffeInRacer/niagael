package dto

import (
	"encoding/json"
	"sort"
	"time"

	"kaffein/product-service/internal/domain"
	"kaffein/product-service/utils"
)

type ProductDetailResponse struct {
	Id              string                 `json:"id"`
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug"`
	Description     string                 `json:"description"`
	Price           int64                  `json:"price"`
	FinalPrice      int64                  `json:"final_price"`
	DiscountPercent int32                  `json:"discount_percent"`
	Stock           int64                  `json:"stock,omitempty"`
	Images          []ProductImageResponse `json:"images"`
	Variants        *VariantData           `json:"variants,omitempty"`
	Category        CategoryBriefResponse  `json:"category"`
	IsFlashSale     bool                   `json:"is_flash_sale"`
	IsActive        bool                   `json:"is_active"`
	CreatedAt       time.Time              `json:"created_at"`
}

type VariantData struct {
	Attributes []string              `json:"attributes"`
	Items      []VariantItemResponse `json:"items"`
}

type VariantItemResponse struct {
	Id              string          `json:"id"`
	Name            string          `json:"name"`
	Price           int64           `json:"price"`
	FinalPrice      int64           `json:"final_price"`
	DiscountPercent int32           `json:"discount_percent"`
	Stock           int64           `json:"stock"`
	Options         json.RawMessage `json:"options"`
}

type ProductImageResponse struct {
	Id       string `json:"id"`
	FileName string `json:"file_name"`
	Sort     int    `json:"sort_order"`
}

type CategoryBriefResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ProductListResponse struct {
	Id              string                `json:"id"`
	Name            string                `json:"name"`
	Slug            string                `json:"slug"`
	Description     string                `json:"description"`
	Price           int64                 `json:"price"`
	FinalPrice      int64                 `json:"final_price"`
	Discount        int64                 `json:"discount"`
	DiscountPercent int32                 `json:"discount_percent"`
	HasVariants     bool                  `json:"has_variants"`
	Image           *ProductImageResponse `json:"image,omitempty"`
	Category        string                `json:"category"`
	IsFlashSale     bool                  `json:"is_flash_sale"`
}

type AdminProductListResponse struct {
	Id          string                `json:"id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	Category    string                `json:"category"`
	HasVariants bool                  `json:"has_variants"`
	Image       *ProductImageResponse `json:"image,omitempty"`
	IsFlashSale bool                  `json:"is_flash_sale"`
}

type VariantFlashSale struct {
	VariantId       string
	DiscountPercent int32
}

func ToProductDetailResponse(p *domain.Product, flashSale *FlashSaleInfo, variantFlashSales []VariantFlashSale) ProductDetailResponse {
	finalPrice := p.Price
	discountPercent := int32(0)
	isFlashSale := false

	if len(p.Variants) == 0 && flashSale != nil && flashSale.DiscountPercent > 0 {
		discountPercent = flashSale.DiscountPercent
		finalPrice = utils.DiscountedPrice(p.Price, discountPercent)
		isFlashSale = true
	}

	resp := ProductDetailResponse{
		Id:              p.Id,
		Name:            p.Name,
		Slug:            p.Slug,
		Description:     p.Description,
		Price:           p.Price,
		FinalPrice:      finalPrice,
		DiscountPercent: discountPercent,
		IsActive:        p.IsActive,
		IsFlashSale:     isFlashSale,
		CreatedAt:       p.CreatedAt,
		Category: CategoryBriefResponse{
			Id:   p.CategoryId,
			Name: p.CategoryName,
		},
	}

	if len(p.Images) > 0 {
		for _, img := range p.Images {
			resp.Images = append(resp.Images, ProductImageResponse{
				Id:       img.Id,
				FileName: img.FileName,
				Sort:     img.SortOrder,
			})
		}
	}

	if len(p.Variants) == 0 {
		resp.Stock = p.Stock - p.StockReserved
	}

	if len(p.Variants) > 0 {
		attributeSet := make(map[string]bool)
		for _, v := range p.Variants {
			var attrs map[string]interface{}
			if err := json.Unmarshal(v.Attributes, &attrs); err == nil {
				for key := range attrs {
					attributeSet[key] = true
				}
			}
		}

		var attributes []string
		for key := range attributeSet {
			attributes = append(attributes, key)
		}
		sort.Strings(attributes)

		flashSaleMap := make(map[string]int32)
		for _, fs := range variantFlashSales {
			flashSaleMap[fs.VariantId] = fs.DiscountPercent
		}
		if len(flashSaleMap) > 0 {
			resp.IsFlashSale = true
		}

		var items []VariantItemResponse
		for _, v := range p.Variants {
			finalPrice := v.Price
			discountPercent := int32(0)
			if percent, ok := flashSaleMap[v.Id]; ok && percent > 0 {
				discountPercent = percent
				finalPrice = utils.DiscountedPrice(v.Price, percent)
			}
			items = append(items, VariantItemResponse{
				Id:              v.Id,
				Name:            v.Name,
				Price:           v.Price,
				FinalPrice:      finalPrice,
				DiscountPercent: discountPercent,
				Stock:           v.Stock - v.StockReserved,
				Options:         v.Attributes,
			})
		}

		resp.Variants = &VariantData{
			Attributes: attributes,
			Items:      items,
		}
	}

	return resp
}

type FlashSaleInfo struct {
	ProductId       string
	DiscountPercent int32
}

func ToProductListResponse(p *domain.Product, flashSale *FlashSaleInfo, variantFlashSales []VariantFlashSale) ProductListResponse {
	price := p.Price
	finalPrice := p.Price
	discountPercent := int32(0)
	hasVariants := len(p.Variants) > 0
	isFlashSale := false

	if hasVariants {
		flashSaleMap := make(map[string]int32, len(variantFlashSales))
		for _, fs := range variantFlashSales {
			flashSaleMap[fs.VariantId] = fs.DiscountPercent
		}

		for i, v := range p.Variants {
			variantPercent := flashSaleMap[v.Id]
			variantFinalPrice := utils.DiscountedPrice(v.Price, variantPercent)
			if i == 0 || variantFinalPrice < finalPrice {
				price = v.Price
				finalPrice = variantFinalPrice
				discountPercent = variantPercent
			}
		}
		isFlashSale = discountPercent > 0
	} else if flashSale != nil && flashSale.DiscountPercent > 0 {
		discountPercent = flashSale.DiscountPercent
		finalPrice = utils.DiscountedPrice(price, discountPercent)
		isFlashSale = true
	}

	discount := price - finalPrice

	resp := ProductListResponse{
		Id:              p.Id,
		Name:            p.Name,
		Slug:            p.Slug,
		Description:     p.Description,
		Price:           price,
		FinalPrice:      finalPrice,
		Discount:        discount,
		DiscountPercent: discountPercent,
		HasVariants:     hasVariants,
		Category:        p.CategoryName,
		IsFlashSale:     isFlashSale,
	}

	if len(p.Images) > 0 {
		img := p.Images[0]
		resp.Image = &ProductImageResponse{
			Id:       img.Id,
			FileName: img.FileName,
			Sort:     img.SortOrder,
		}
	}

	return resp
}

func ToAdminProductListResponse(p *domain.Product, flashSale *FlashSaleInfo) AdminProductListResponse {
	hasVariants := len(p.Variants) > 0
	isFlashSale := false

	if flashSale != nil && flashSale.DiscountPercent > 0 {
		isFlashSale = true
	}

	resp := AdminProductListResponse{
		Id:          p.Id,
		Name:        p.Name,
		Slug:        p.Slug,
		Category:    p.CategoryName,
		HasVariants: hasVariants,
		IsFlashSale: isFlashSale,
	}

	if len(p.Images) > 0 {
		img := p.Images[0]
		resp.Image = &ProductImageResponse{
			Id:       img.Id,
			FileName: img.FileName,
			Sort:     img.SortOrder,
		}
	}

	return resp
}
