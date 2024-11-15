package config

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateConfig(config *Config) error {
	validate := validator.New()
	validate.RegisterStructValidation(validateShardCount, Config{})

	if err := validate.Struct(config); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return fmt.Errorf("unexpected validation error type: %v", err)
		}

		processedFields := make(map[string]bool)

		for _, validationErr := range validationErrors {
			fieldName := strings.TrimPrefix(validationErr.Namespace(), "Config.")

			if processedFields[fieldName] {
				continue
			}
			processedFields[fieldName] = true

			log.Printf("Field %s failed validation: %s", fieldName, validationErr.Tag())

			if err := applyDefaultValuesFromTags(config, fieldName); err != nil {
				return fmt.Errorf("error setting default for %s: %v", fieldName, err)
			}
		}
	}
	return nil
}

func validateShardCount(sl validator.StructLevel) {
	config := sl.Current().Interface().(Config)
	if config.Performance.NumShards <= 0 && config.Performance.NumShards != -1 {
		sl.ReportError(config.Performance.NumShards, "NumShards", "NumShards", "invalidValue", "must be -1 or greater than 0")
	}
}

func applyDefaultValuesFromTags(config *Config, fieldName string) error {
	configType := reflect.TypeOf(config).Elem()
	configValue := reflect.ValueOf(config).Elem()

	// Split the field name if it refers to a nested struct
	parts := strings.Split(fieldName, ".")
	var field reflect.StructField
	var fieldValue reflect.Value
	var found bool

	// Traverse the struct to find the nested field
	for i, part := range parts {
		// If it's the first field, just look in the top-level struct
		if i == 0 {
			field, found = configType.FieldByName(part)
			if !found {
				log.Printf("Warning: %s field not found", part)
				return fmt.Errorf("field %s not found in config struct", part)
			}
			fieldValue = configValue.FieldByName(part)
		} else {
			// Otherwise, the struct is nested, so navigate into it
			if fieldValue.Kind() == reflect.Struct {
				field, found = fieldValue.Type().FieldByName(part)
				if !found {
					log.Printf("Warning: %s field not found in %s", part, fieldValue.Type())
					return fmt.Errorf("field %s not found in struct %s", part, fieldValue.Type())
				}
				fieldValue = fieldValue.FieldByName(part)
			} else {
				log.Printf("Warning: %s is not a struct", fieldName)
				return fmt.Errorf("%s is not a struct", fieldName)
			}
		}
	}

	defaultValue := field.Tag.Get("default")
	if defaultValue == "" {
		log.Printf("Warning: %s field has no default value to set, leaving empty string", fieldName)
		return nil
	}

	if err := setField(fieldValue, defaultValue); err != nil {
		return fmt.Errorf("error setting default value for %s: %v", fieldName, err)
	}

	log.Printf("Setting default value for %s to: %s", fieldName, defaultValue)
	return nil
}
