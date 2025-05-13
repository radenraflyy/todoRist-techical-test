package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func CustomErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' wajib diisi.", fe.Field())
	case "min":
		return fmt.Sprintf("Field '%s' harus memiliki minimal %s karakter.", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' harus memiliki maksimal %s karakter.", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("Field '%s' harus memiliki format email.", fe.Field())
	default:
		return fmt.Sprintf("Field '%s' tidak valid.", fe.Field())
	}
}
