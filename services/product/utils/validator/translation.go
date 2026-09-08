package validator

import (
	"kaffein/product-service/utils/constants"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func customTranslation(v *validator.Validate, trans ut.Translator) {
	customTranslations := map[string]string{
		"required":       constants.ErrValidationRequired,
		"numeric":        constants.ErrValidationNumeric,
		"image":          constants.ErrValidationImage,
		"image_size_max": constants.ErrValidationImageSize,
		"datetime":       constants.ErrValidationDatetime,
		"min":            constants.ErrValidationMin,
		"max":            constants.ErrValidationMax,
		"gte":            constants.ErrValidationGTE,
		"gt":             constants.ErrValidationGT,
		"lte":            constants.ErrValidationLTE,
		"lt":             constants.ErrValidationLT,
		"len":            constants.ErrValidationLen,
		"email":          constants.ErrValidationEmail,
		"url":            constants.ErrValidationURL,
		"boolean":        constants.ErrValidationBoolean,
	}

	for tag, msg := range customTranslations {
		tag, msg := tag, msg
		_ = v.RegisterTranslation(tag, trans,
			func(ut ut.Translator) error { return ut.Add(tag, msg, true) },
			func(ut ut.Translator, fe validator.FieldError) string {
				translated, _ := ut.T(tag, fe.Field(), fe.Param())
				return translated
			},
		)
	}
}
