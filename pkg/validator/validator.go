package validator

import (
	"fmt"

	v "github.com/go-playground/validator/v10"
	"ns-tracking-go/pkg/errorx"
)

var validate = v.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Struct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	verrs, ok := err.(v.ValidationErrors)
	if !ok {
		return errorx.NewError(errorx.InvalidParameter)
	}

	errs := make([]ValidationError, 0, len(verrs))
	for _, e := range verrs {
		errs = append(errs, ValidationError{
			Field:   e.Field(),
			Message: fmt.Sprintf("validation failed on '%s' tag", e.Tag()),
		})
	}

	return errorx.NewCodeError(errorx.InvalidParameter, formatErrors(errs))
}

func formatErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "validation failed"
	}
	s := errs[0].Field + ": " + errs[0].Message
	for i := 1; i < len(errs); i++ {
		s += "; " + errs[i].Field + ": " + errs[i].Message
	}
	return s
}
