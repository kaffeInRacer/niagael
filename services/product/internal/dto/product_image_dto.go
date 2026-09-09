package dto

type UploadProductImageDto struct {
	SortOrder int32 `form:"sort_order" validate:"clamp=0 100"`
}
