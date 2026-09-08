package validator

import (
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateClamp(fl validator.FieldLevel) bool {
	param := fl.Param()
	if param == "" {
		return false
	}

	parts := strings.Fields(param)

	var min, max int64
	var err error

	switch len(parts) {
	case 1:
		// Single value: "10" means min=10, max=MaxInt
		min, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return false
		}
		max = int64(^uint64(0) >> 1) // MaxInt64
	case 2:
		// Two values: "10-100"
		min, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return false
		}
		max, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return false
		}
	default:
		return false
	}

	field := fl.Field()
	if !field.CanSet() {
		return field.Int() >= min && field.Int() <= max
	}

	val := field.Int()
	switch {
	case val < min:
		field.SetInt(min)
	case val > max:
		field.SetInt(max)
	}

	return true
}
