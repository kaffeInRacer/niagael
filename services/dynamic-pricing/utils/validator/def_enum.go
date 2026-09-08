package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func validateDefEnum(fl validator.FieldLevel) bool {
	param := fl.Param()
	if param == "" {
		return false
	}

	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}

	val := field.String()

	start := 0
	var firstEnum string
	found := false

	for i := 0; i <= len(param); i++ {
		if i == len(param) || param[i] == ' ' {
			if start < i {
				current := param[start:i]
				if firstEnum == "" {
					firstEnum = current
				}
				if val == current {
					found = true
					break
				}
			}
			start = i + 1
		}
	}

	if found {
		return true
	}

	if field.CanSet() && firstEnum != "" {
		field.SetString(firstEnum)
		return true
	}

	return false
}
