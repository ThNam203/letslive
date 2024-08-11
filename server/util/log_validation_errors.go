package util

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func LogValidationErrors(err error) {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		fmt.Println(err)
		return
	}
	for _, err := range err.(validator.ValidationErrors) {
		fmt.Println(err.Namespace())
		fmt.Println(err.Field())
		fmt.Println(err.StructNamespace())
		fmt.Println(err.StructField())
		fmt.Println(err.Tag())
		fmt.Println(err.ActualTag())
		fmt.Println(err.Kind())
		fmt.Println(err.Type())
		fmt.Println(err.Value())
		fmt.Println(err.Param())
		fmt.Println()
	}
}
