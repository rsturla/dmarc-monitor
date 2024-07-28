package main

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
)

// Extracts the tag from a plus-addressed email
func extractPlusAddressTag(email string) (string, error) {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("invalid email address: %s", email)
	}

	parts := strings.Split(address.Address, "@")

	localParts := strings.Split(parts[0], "+")
	if len(localParts) < 2 {
		return "", nil
	}

	return localParts[1], nil
}

// Retrieves the raw email location from S3
func getRawEmailLocation(ctx context.Context, awsClient *AWSClient, bucket, prefix, messageId string) (string, error) {
	messageLocation := fmt.Sprintf("%s%s", prefix, messageId)

	// Call HeadObject API to check if the object exists
	exists, err := awsClient.s3ObjectExists(ctx, bucket, messageLocation)
	if err != nil {
		return "", fmt.Errorf("error checking if object exists: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("object does not exist: %s", messageLocation)
	}

	return fmt.Sprintf("s3://%s/%s", bucket, messageLocation), nil
}
