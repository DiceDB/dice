// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ConfigParser handles the parsing of configuration files
type ConfigParser struct {
	// store holds the raw key-value pairs from the config file
	store map[string]string
}

// NewConfigParser creates a new instance of ConfigParser
func NewConfigParser() *ConfigParser {
	return &ConfigParser{
		store: make(map[string]string),
	}
}

// ParseFromFile reads the configuration data from a file
func (p *ConfigParser) ParseFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	return processConfigData(scanner, p)
}

// ParseFromStdin reads the configuration data from stdin
func (p *ConfigParser) ParseFromStdin() error {
	scanner := bufio.NewScanner(os.Stdin)
	return processConfigData(scanner, p)
}

// ParseDefaults populates a struct with default values based on struct tag `default`
func (p *ConfigParser) ParseDefaults(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	return p.unmarshalStruct(val, "")
}

// Loadconfig populates a struct with configuration values based on struct tags
func (p *ConfigParser) Loadconfig(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	if err := p.unmarshalStruct(val, ""); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(DiceConfig); err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	return nil
}

// processConfigData reads the configuration data line by line and stores it in the ConfigParser
func processConfigData(scanner *bufio.Scanner, p *ConfigParser) error {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			slog.Warn("invalid config line", slog.String("line", line))
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		p.store[key] = value
	}

	return scanner.Err()
}

// unmarshalStruct handles the recursive struct parsing.
func (p *ConfigParser) unmarshalStruct(val reflect.Value, prefix string) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields just like how encoding/json does
		if !field.CanSet() {
			continue
		}

		// Get config key or field name
		key := fieldType.Tag.Get("config")

		// Use field name as key if not specified in tag
		if key == "" {
			key = strings.ToLower(fieldType.Name)
		}

		// Skip fields with "-" tag
		if key == "-" {
			continue
		}

		// Apply nested struct's tag as prefix
		fullKey := key
		if prefix != "" {
			fullKey = fmt.Sprintf("%s.%s", prefix, key)
		}

		// Recursively process nested structs with their prefix
		if field.Kind() == reflect.Struct {
			if err := p.unmarshalStruct(field, fullKey); err != nil {
				return err
			}
			continue
		}

		// Fetch and set value for non-struct fields
		value, exists := p.store[fullKey]
		if !exists {
			// Use default value from tag if available
			if defaultValue := fieldType.Tag.Get("default"); defaultValue != "" {
				value = defaultValue
			} else {
				continue
			}
		}

		if err := setField(field, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", fullKey, err)
		}
	}

	return nil
}

// setField sets the appropriate field value based on its type
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration type
			duration, err := parseDuration(value)
			if err != nil {
				return fmt.Errorf("failed to parse duration: %w", err)
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			// Handle other integer types
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse integer: %w", err)
			}
			field.SetInt(val)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse unsigned integer: %w", err)
		}
		field.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float: %w", err)
		}
		field.SetFloat(val)

	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse boolean: %w", err)
		}
		field.SetBool(val)

	case reflect.Slice:
		// Handle slices of basic types
		elemType := field.Type().Elem()
		values := strings.Split(value, ",")
		slice := reflect.MakeSlice(field.Type(), len(values), len(values))
		for i, v := range values {
			elem := slice.Index(i)
			elemVal := reflect.New(elemType).Elem()
			if err := setField(elemVal, strings.TrimSpace(v)); err != nil {
				return fmt.Errorf("failed to parse slice element at index %d: %w", i, err)
			}
			elem.Set(elemVal)
		}
		field.Set(slice)

	default:
		return fmt.Errorf("unsupported type: %s", field.Type())
	}

	return nil
}

// parseDuration handles parsing of time.Duration with proper validation.
func parseDuration(value string) (time.Duration, error) {
	if value == "" {
		return 0, fmt.Errorf("duration string is empty")
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %s", value)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive, got: %s", value)
	}
	return duration, nil
}
