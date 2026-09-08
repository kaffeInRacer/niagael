package dto

type CreateProductDto struct {
	Id          string
	CategoryId  string `json:"category_id" validate:"required,uuid"`
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string
	Description string `json:"description"`
	Price       int64  `json:"price" validate:"gte=0"`
	Stock       int64  `json:"stock" validate:"gte=0"`
	IsActive    bool   `json:"is_active"`
}

type UpdateProductDto struct {
	Id          string
	CategoryId  string `json:"category_id" validate:"required,uuid"`
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string
	Description string `json:"description"`
	Price       int64  `json:"price" validate:"gte=0"`
	Stock       int64  `json:"stock" validate:"gte=0"`
	IsActive    bool   `json:"is_active"`
}

type CreateProductImageDto struct {
	Id        string
	ProductId string
	FileName  string
	SortOrder int
}
