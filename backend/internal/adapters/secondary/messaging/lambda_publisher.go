package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaPublisher invokes the search worker Lambda directly (no SQS).
// The call is fire-and-forget using InvocationType "Event" (async).
type LambdaPublisher struct {
	client       *lambda.Client
	functionName string
}

// NewLambdaPublisher creates a Lambda client using standard AWS SDK config.
func NewLambdaPublisher(ctx context.Context, functionName string, awsRegion string) (*LambdaPublisher, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := lambda.NewFromConfig(cfg)

	return &LambdaPublisher{
		client:       client,
		functionName: functionName,
	}, nil
}

// PublishSyncEvent invokes the Lambda function asynchronously with the event payload.
// Uses InvocationType "Event" so the Go server does NOT wait for Lambda to finish.
func (p *LambdaPublisher) PublishSyncEvent(ctx context.Context, event domain.SearchSyncEvent) error {
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal search sync event: %w", err)
	}

	go func() {
		_, err := p.client.Invoke(context.Background(), &lambda.InvokeInput{
			FunctionName:   &p.functionName,
			Payload:        payloadBytes,
			InvocationType: "Event", // Async: Lambda queues it internally, Go doesn't block
		})
		if err != nil {
			log.Printf("⚠️ Failed to invoke search Lambda: %v", err)
		}
	}()

	return nil
}

// GenerateEmbedding invokes the Lambda synchronously to vectorize a search query.
func (p *LambdaPublisher) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"action": "embed_query",
		"data": map[string]string{
			"text": text,
		},
	}
	
	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed_query event: %w", err)
	}

	res, err := p.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   &p.functionName,
		Payload:        payloadBytes,
		InvocationType: "RequestResponse", // Synchronous block
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke lambda for embedding: %w", err)
	}

	if res.FunctionError != nil {
		return nil, fmt.Errorf("lambda returned error: %s", *res.FunctionError)
	}

	// The Lambda returns a JSON like: {"statusCode": 200, "body": "{\"vector\": [0.1, 0.2, ...]}"}
	var outerResp struct {
		StatusCode int             `json:"statusCode"`
		Body       json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal(res.Payload, &outerResp); err != nil {
		return nil, fmt.Errorf("failed to parse outer lambda response: %w", err)
	}

	if outerResp.StatusCode != 200 {
		return nil, fmt.Errorf("lambda failed with status %d: %s", outerResp.StatusCode, string(outerResp.Body))
	}

	// Because of potential double JSON-encoding from Python's json.dumps, 
	// outerResp.Body might be a JSON string like `"{\"vector\":...}"` instead of an unescaped string `{"vector":...}`.
	// Let's unmarshal it into a standard string first to unescape it, if it starts with a quote.
	var bodyStr string
	if len(outerResp.Body) > 0 && outerResp.Body[0] == '"' {
		if err := json.Unmarshal(outerResp.Body, &bodyStr); err != nil {
			return nil, fmt.Errorf("failed to unescape body string: %w", err)
		}
	} else {
		// If it's not a JSON string, assume it's already an unescaped raw JSON object
		bodyStr = string(outerResp.Body)
	}

	var innerResp struct {
		Vector []float32 `json:"vector"`
	}
	if err := json.Unmarshal([]byte(bodyStr), &innerResp); err != nil {
		return nil, fmt.Errorf("failed to parse inner lambda response: %w. Raw body: %s", err, bodyStr)
	}

	return innerResp.Vector, nil
}
