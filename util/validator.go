package util

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type validationError struct {
	Namespace       string `json:"namespace"`
	Field           string `json:"field"`
	StructNamespace string `json:"structNamespace"`
	StructField     string `json:"structField"`
	Tag             string `json:"tag"`
	ActualTag       string `json:"actualTag"`
	Kind            string `json:"kind"`
	Type            string `json:"type"`
	Value           string `json:"value"`
	Param           string `json:"param"`
	Message         string `json:"message"`
}

// JoinValidationErrors combines validation errors together as one error
func JoinValidationErrors(e error) error {
	if errs, ok := e.(validator.ValidationErrors); ok {
		joined := make([]error, 0, len(errs))

		for _, err := range errs {
			e := validationError{
				Namespace:       err.Namespace(),
				Field:           err.Field(),
				StructNamespace: err.StructNamespace(),
				StructField:     err.StructField(),
				Tag:             err.Tag(),
				ActualTag:       err.ActualTag(),
				Kind:            fmt.Sprintf("%v", err.Kind()),
				Type:            fmt.Sprintf("%v", err.Type()),
				Value:           fmt.Sprintf("%v", err.Value()),
				Param:           err.Param(),
				Message:         err.Error(),
			}

			indent, mErr := json.MarshalIndent(e, "", "  ")
			if mErr != nil {
				fmt.Println(mErr)
				panic(mErr)
			}

			joined = append(joined, errors.New(string(indent)))
		}

		return errors.Join(joined...)
	}

	return e
}
