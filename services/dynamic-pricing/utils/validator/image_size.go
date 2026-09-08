package validator

import (
	"mime/multipart"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func validateImageSizeMax(fl validator.FieldLevel) bool {
	var files []*multipart.FileHeader

	switch v := fl.Field().Interface().(type) {
	case *multipart.FileHeader:
		if v != nil {
			files = append(files, v)
		}
	case []*multipart.FileHeader:
		files = v
	}

	if len(files) == 0 {
		return false
	}

	// Ambil angka dari tag (misal "1000") lalu konversi KB ke Bytes
	maxKB, err := strconv.ParseInt(fl.Param(), 10, 64)
	if err != nil || maxKB <= 0 {
		return false
	}
	maxBytes := maxKB * 1024

	for _, file := range files {
		if file.Size > maxBytes {
			return false
		}
	}

	return true
}
