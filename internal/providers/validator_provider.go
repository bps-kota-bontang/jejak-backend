package providers

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// NewValidator provides a new instance of a Validator.
// Validation error field names are taken from the json tag so that
// clients receive snake_case keys (e.g. "old_password") instead of
// the Go struct name (e.g. "OldPassword").
func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}
