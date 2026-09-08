package domain

type ProductImage struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id"`
	FileName  string `json:"file_name"`
	SortOrder int    `json:"sort_order"`
}
