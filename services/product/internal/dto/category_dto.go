package dto

type CreateCategoryDto struct {
	Id          string
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string
	Description string `json:"description"`
}

type UpdateCategoryDto struct {
	Id          string
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string
	Description string `json:"description"`
}
