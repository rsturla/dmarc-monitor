package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (c *AWSClient) DynamoDBPutItem(ctx context.Context, tableName string, item *map[string]dynamodbTypes.AttributeValue) error {
	_, err := c.DynamoDb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      *item,
	})
	if err != nil {
		return fmt.Errorf("error putting item to DynamoDB table %s: %w", tableName, err)
	}

	return nil
}

func (c *AWSClient) DynamoDBPutBatchItems(ctx context.Context, tableName string, items []map[string]dynamodbTypes.AttributeValue) error {
	const batchSize = 25

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		// Slice the items to create a batch of 25 (or fewer for the last batch)
		batch := items[i:end]

		var writeRequests []dynamodbTypes.WriteRequest
		for _, item := range batch {
			// Check if item is not empty and contains required fields
			if len(item) == 0 {
				return fmt.Errorf("missing required fields in item: %v", item)
			}

			writeRequests = append(writeRequests, dynamodbTypes.WriteRequest{
				PutRequest: &dynamodbTypes.PutRequest{
					Item: item,
				},
			})
		}

		if len(writeRequests) == 0 {
			return fmt.Errorf("no valid items to write in batch starting at index %d", i)
		}

		_, err := c.DynamoDb.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]dynamodbTypes.WriteRequest{
				tableName: writeRequests,
			},
		})
		if err != nil {
			return fmt.Errorf("error putting batch items to DynamoDB table %s (batch starting at index %d): %w", tableName, i, err)
		}
	}

	return nil
}
