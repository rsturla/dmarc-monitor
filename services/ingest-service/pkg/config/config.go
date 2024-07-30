package config

import (
	"fmt"
	"os"
	"reflect"
)

type Config struct {
	ReportStorageBucketName string `env:"REPORT_STORAGE_BUCKET_NAME"`
	RawEmailQueueURL        string `env:"RAW_EMAIL_QUEUE_URL"`
	AttachmentQueueURL      string `env:"ATTACHMENT_QUEUE_URL"`
}

// NewConfig retrieves all required environment variables and returns a Config struct.
func NewConfig() (*Config, error) {
	config := &Config{}
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
