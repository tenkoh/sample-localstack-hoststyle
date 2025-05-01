package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tenkoh/sample-localstack-hoststyle/presigner"
)

type PresignEvent struct {
	Key string `json:"key"`
}

type PresignResponse struct {
	URL string `json:"url"`
}

func handler(client *s3.PresignClient, bucket string) func(ctx context.Context, event json.RawMessage) (PresignResponse, error) {
	p := presigner.NewPresigner(client, bucket)
	return func(ctx context.Context, event json.RawMessage) (PresignResponse, error) {
		var presignEvent PresignEvent
		if err := json.Unmarshal(event, &presignEvent); err != nil {
			return PresignResponse{}, err
		}

		presignedURL, err := p.PresignedURL(ctx, presignEvent.Key)
		if err != nil {
			return PresignResponse{}, err
		}

		return PresignResponse{URL: presignedURL}, nil
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	client := s3.NewPresignClient(s3.NewFromConfig(cfg))
	lambda.Start(handler(client, os.Getenv("BUCKET_NAME")))
}
