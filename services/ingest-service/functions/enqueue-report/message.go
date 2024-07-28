package main

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
func getRawEmailLocation(ctx context.Context, s3Client *s3.Client, bucket, prefix, messageId string) (string, error) {
	messageLocation := fmt.Sprintf("%s%s", prefix, messageId)

	// Call HeadObject API to check if the object exists
	_, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &messageLocation,
	})
	if err != nil {
		return "", fmt.Errorf("error getting raw email location for message ID %s: %w", messageId, err)
	}

	return fmt.Sprintf("s3://%s/%s", bucket, messageLocation), nil
}
