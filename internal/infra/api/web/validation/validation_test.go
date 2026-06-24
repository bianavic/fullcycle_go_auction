package validation_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"fullcycle-auction_go/internal/infra/api/web/validation"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestValidateUUID(t *testing.T) {
	t.Parallel()

	t.Run("valid UUID returns nil", func(t *testing.T) {
		t.Parallel()
		err := validation.ValidateUUID(uuid.NewString(), "auctionId")
		require.Nil(t, err)
	})

	t.Run("invalid UUID returns bad request", func(t *testing.T) {
		t.Parallel()
		err := validation.ValidateUUID("not-a-uuid", "auctionId")
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.Code)
		require.Len(t, err.Causes, 1)
		require.Equal(t, "auctionId", err.Causes[0].Field)
	})

	t.Run("empty string returns bad request", func(t *testing.T) {
		t.Parallel()
		err := validation.ValidateUUID("", "userId")
		require.NotNil(t, err)
		require.Equal(t, "userId", err.Causes[0].Field)
	})
}

func TestValidateErr(t *testing.T) {
	t.Parallel()

	t.Run("UnmarshalTypeError returns bad request with field info", func(t *testing.T) {
		t.Parallel()
		typeErr := &json.UnmarshalTypeError{
			Value: "string",
			Type:  reflect.TypeOf(0),
			Field: "price",
		}
		result := validation.ValidateErr(typeErr)
		require.NotNil(t, result)
		require.Equal(t, http.StatusBadRequest, result.Code)
		require.Len(t, result.Causes, 1)
		require.Equal(t, "price", result.Causes[0].Field)
		require.Contains(t, result.Causes[0].Message, "int")
	})

	t.Run("ValidationErrors returns bad request with translated causes", func(t *testing.T) {
		t.Parallel()
		validate := validator.New()
		input := struct {
			Name string `validate:"required"`
		}{}
		err := validate.Struct(input)
		require.Error(t, err)

		result := validation.ValidateErr(err)
		require.NotNil(t, result)
		require.Equal(t, http.StatusBadRequest, result.Code)
		require.NotEmpty(t, result.Causes)
	})

	t.Run("unknown error returns generic bad request", func(t *testing.T) {
		t.Parallel()
		result := validation.ValidateErr(fmt.Errorf("some parsing error"))
		require.NotNil(t, result)
		require.Equal(t, http.StatusBadRequest, result.Code)
		require.Equal(t, "error trying to convert fields", result.Message)
	})
}
