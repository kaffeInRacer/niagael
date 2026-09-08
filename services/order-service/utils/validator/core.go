package validator

import (
	"errors"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en2 "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	validate   *validator.Validate
	translator ut.Translator
}

var (
	once    sync.Once
	instance *Validator
)

func Get() *Validator {
	once.Do(func() {
		instance = newValidator()
	})
	return instance
}

func newValidator() *Validator {
	locale := en.New()
	uni := ut.New(locale)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()

	_ = en2.RegisterDefaultTranslations(validate, trans)
	_ = validate.RegisterValidation("clamp", validateClamp)
	_ = validate.RegisterValidation("image", validateImage)
	_ = validate.RegisterValidation("image_size_max", validateImageSizeMax)
	_ = validate.RegisterValidation("def_enum", validateDefEnum)

	customTranslation(validate, trans)

	return &Validator{
		validate:   validate,
		translator: trans,
	}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

func (v *Validator) ValidateStruct(i interface{}) map[string]string {
	if err := v.validate.Struct(i); err != nil {
		return v.Translate(err)
	}
	return nil
}

func (v *Validator) Translate(err error) map[string]string {
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	ok := errors.As(err, &errs)
	if !ok {
		return nil
	}

	translations := make(map[string]string)
	for _, e := range errs {
		translations[e.Field()] = e.Translate(v.translator)
	}

	return translations
}
