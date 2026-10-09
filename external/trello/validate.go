package trello

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	playground "github.com/go-playground/validator/v10"
)

const trelloIDRegex = `^[0-9a-fA-F]{24}$`

var validator = newStructValidator()

func validate(value any) error {
	err := validator.Struct(value)
	if err == nil {
		return nil
	}

	validationErrors, ok := errors.AsType[playground.ValidationErrors](err)
	if !ok || len(validationErrors) == 0 {
		return err
	}
	return fieldError{validationErrors[0]}
}

func newStructValidator() *playground.Validate {
	validate := playground.New(playground.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("trelloID", ValidateTrelloID); err != nil {
		panic(err)
	}
	if err := validate.RegisterValidation("labelColor", ValidateLabelColor); err != nil {
		panic(err)
	}
	return validate
}

func ValidateTrelloID(field playground.FieldLevel) bool {
	match, err := regexp.MatchString(trelloIDRegex, field.Field().String())
	if !match || err != nil {
		return false
	}
	return true
}

func ValidateLabelColor(field playground.FieldLevel) bool {
	return slices.Contains(getLabelColors(), field.Field().String())
}

type fieldError struct {
	playground.FieldError
}

func (err fieldError) Error() string {
	field := err.Field()
	switch err.Tag() {
	case "trelloID":
		return fmt.Sprintf("%s is not a valid Trello ID", err.Param())
	case "labelColor":
		return fmt.Sprintf("%s is not a valid label color", err.Param())
	default:
		return field + " is invalid"
	}
}
