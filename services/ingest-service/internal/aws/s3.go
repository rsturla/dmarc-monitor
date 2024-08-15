package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3ObjectExists checks if an object exists in an S3 bucket.
func (c *AWSClient) S3ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := c.S3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		var notFoundErr *s3Types.NotFound
		if errors.As(err, &notFoundErr) {
			return false, nil
		}
		return false, err // Return the actual error if it's not a NotFound error
	}
	return true, nil
}

// S3GetObject retrieves an object from an S3 bucket. The object is returned as a byte slice.
func (c *AWSClient) S3GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := c.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3: %w", err)
	}
	defer obj.Body.Close()

	return io.ReadAll(obj.Body)
}

func (c *AWSClient) S3PutObject(ctx context.Context, bucket, key string, contentType string, body []byte) error {
	_, err := c.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(body),
	})
	if err != nil {
		return fmt.Errorf("error putting object to S3: %w", err)
	}
	return nil
}
