package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSPublisher struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSPublisher configures the AWS SQS client.
// It relies on aws-sdk-go-v2/config which automatically loads AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_REGION from the environment.
func NewSQSPublisher(ctx context.Context, queueURL string, awsRegion string) (*SQSPublisher, error) {
	// Standard config load mechanism for aws-sdk-v2
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

	return &SQSPublisher{
		client:   client,
		queueURL: queueURL,
	}, nil
}

// PublishSyncEvent sends the event payload as a JSON message to AWS SQS.
func (p *SQSPublisher) PublishSyncEvent(ctx context.Context, event domain.SearchSyncEvent) error {
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal search sync event: %w", err)
	}

	payloadStr := string(payloadBytes)

	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(payloadStr),
	})

	if err != nil {
		return fmt.Errorf("failed to publish to SQS: %w", err)
	}

	return nil
}
