package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	ssmClient     *ssm.Client
	ssmClientOnce sync.Once
)

// NewConfig retrieves all required environment variables and returns a Config struct.
func NewConfig[T any]() (*T, error) {
	config := new(T)
	v := reflect.ValueOf(config).Elem()
	typeOfConfig := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		envVarName := typeOfConfig.Field(i).Tag.Get("env")
		ssmVarName := typeOfConfig.Field(i).Tag.Get("ssm")

		var value string
		var err error
		switch {
		case ssmVarName != "":
			value, err = getSSMParameter(ssmVarName)
		case envVarName != "":
			value, err = getEnvironmentVariable(envVarName)
		default:
			continue
		}
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

// getSSMParameter fetches the value of an SSM parameter, decrypting it if necessary.
func getSSMParameter(key string) (string, error) {
	// Initialize the SSM client only once
	ssmClientOnce.Do(func() {
		ssmClient = ssm.NewFromConfig(aws.Config{})
	})

	param, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return "", fmt.Errorf("failed to get parameter %s from SSM: %v", key, err)
	}

	return *param.Parameter.Value, nil
}
