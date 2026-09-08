package dto

type ListProductParams struct {
	Search          string  `form:"search" validate:"omitempty,max=100"`
	IsActive        *bool   `form:"is_active" validate:"omitempty"`
	IsPromoExcluded *bool   `form:"is_promo_excluded" validate:"omitempty"`
	CategoryID      string  `form:"category_id" validate:"omitempty,uuid"`
	MinPrice        string  `form:"min_price" validate:"omitempty,numeric"`
	MaxPrice        string  `form:"max_price" validate:"omitempty,numeric"`
	FromDate        *string `form:"from_date" validate:"omitempty,datetime=2006-01-02"`
	ToDate          *string `form:"to_date" validate:"omitempty,datetime=2006-01-02"`
	OrderBy         string  `form:"order_by" validate:"def_enum=created_at name price"`
	OrderDir        string  `form:"order_dir" validate:"def_enum=desc asc"`
	PageSize        int32   `form:"page_size" validate:"clamp=10 100"`
	PageOffset      int32   `form:"page" validate:"clamp=0"`
}
