package validation

import (
	"encoding/json"
	"errors"
	"fullcycle-auction_go/configuration/httperr"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	validatoren "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
)

var (
	translator ut.Translator
)

func init() {
	if value, ok := binding.Validator.Engine().(*validator.Validate); ok {
		en := en.New()
		enTransl := ut.New(en, en)
		translator, _ = enTransl.GetTranslator("en")
		_ = validatoren.RegisterDefaultTranslations(value, translator)
	}
}

func ValidateUUID(value, field string) *httperr.RestErr {
	if err := uuid.Validate(value); err != nil {
		return httperr.NewBadRequestError("Invalid fields", httperr.Causes{
			Field:   field,
			Message: "Invalid UUID value",
		})
	}
	return nil
}

func ValidateErr(validationErr error) *httperr.RestErr {
	if _, ok := errors.AsType[*json.UnmarshalTypeError](validationErr); ok {
		return httperr.NewNotFoundError("Invalid type error")
	} else if jsonValidation, ok := errors.AsType[validator.ValidationErrors](validationErr); ok {
		var errorCauses []httperr.Causes

		for _, e := range jsonValidation {
			errorCauses = append(errorCauses, httperr.Causes{
				Field:   e.Field(),
				Message: e.Translate(translator),
			})
		}

		return httperr.NewBadRequestError("Invalid field values", errorCauses...)
	}
	return httperr.NewBadRequestError("Error trying to convert fields")
}
