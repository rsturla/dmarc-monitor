package config

import (
	"fmt"
	"os"
	"reflect"
)

// NewConfig retrieves all required environment variables and returns a Config struct.
func NewConfig[T any]() (*T, error) {
	config := new(T)
	v := reflect.ValueOf(config).Elem()
	typeOfConfig := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		envVarName := typeOfConfig.Field(i).Tag.Get("env")
		if envVarName == "" {
			envVarName = typeOfConfig.Field(i).Name
		}
		value, err := getEnvironmentVariable(envVarName)
		if err != nil {
			return nil, err
		}
		field.SetString(value)
	}

	return config, nil
}

// getEnvironmentVariable fetches the value of an environment variable.
func getEnvironmentVariable(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}
