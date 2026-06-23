package validation

import (
	"encoding/json"
	"errors"
	"fullcycle-auction_go/configuration/httperr"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	validator_en "github.com/go-playground/validator/v10/translations/en"
)

var (
	Validate   = validator.New()
	translator ut.Translator
)

func init() {
	if value, ok := binding.Validator.Engine().(*validator.Validate); ok {
		en := en.New()
		enTransl := ut.New(en, en)
		translator, _ = enTransl.GetTranslator("en")
		_ = validator_en.RegisterDefaultTranslations(value, translator)
	}
}

func ValidateErr(validation_err error) *httperr.RestErr {
	var jsonErr *json.UnmarshalTypeError
	var jsonValidation validator.ValidationErrors

	if errors.As(validation_err, &jsonErr) {
		return httperr.NewNotFoundError("Invalid type error")
	} else if errors.As(validation_err, &jsonValidation) {
		errorCauses := []httperr.Causes{}

		for _, e := range jsonValidation {
			errorCauses = append(errorCauses, httperr.Causes{
				Field:   e.Field(),
				Message: e.Translate(translator),
			})
		}

		return httperr.NewBadRequestError("Invalid field values", errorCauses...)
	} else {
		return httperr.NewBadRequestError("Error trying to convert fields")
	}
}
