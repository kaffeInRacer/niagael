package dto

type ListCategoryParams struct {
	Search              string `form:"search" validate:"omitempty,max=100"`
	IsActive            *bool  `form:"is_active" validate:"omitempty"`
	OrderBy             string `form:"order_by" validate:"def_enum=created_at name slug"`
	OrderDir            string `form:"order_dir" validate:"def_enum=desc asc"`
	PageSize            int32  `form:"page_size" validate:"clamp=10 100"`
	PageOffset          int32  `form:"page" validate:"clamp=0"`
	IncludeProductCount bool   `form:"include_product_count" validate:"omitempty"`
	ActiveProductsOnly  bool   `form:"-"`
}
