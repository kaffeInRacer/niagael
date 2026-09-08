package validator

import (
	"mime/multipart"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-playground/validator/v10"
)

// Mapping ekstensi (dari tag) ke MIME type
var extToMime = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"bmp":  "image/bmp",
	"webp": "image/webp",
}

func validateImage(fl validator.FieldLevel) bool {
	exts := strings.Fields(fl.Param())
	if len(exts) == 0 {
		return false
	}

	allowedMimes := make(map[string]bool, len(exts))
	for _, ext := range exts {
		if mime, exists := extToMime[strings.ToLower(ext)]; exists {
			allowedMimes[mime] = true
		}
	}
	if len(allowedMimes) == 0 {
		return false
	}

	var files []*multipart.FileHeader
	val := fl.Field().Interface()

	switch v := val.(type) {
	case *multipart.FileHeader:
		if v != nil {
			files = append(files, v)
		}
	case []*multipart.FileHeader:
		files = v
	default:
		return false
	}

	if len(files) == 0 {
		return false
	}

	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			return false
		}

		mtype, err := mimetype.DetectReader(f)
		f.Close()

		if err != nil || !allowedMimes[mtype.String()] {
			return false
		}
	}

	return true
}
