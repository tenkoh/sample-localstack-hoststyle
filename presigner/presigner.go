package presigner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Presigner struct {
	client *s3.PresignClient
	bucket string
}

func NewPresigner(client *s3.PresignClient, bucket string) *Presigner {
	return &Presigner{client: client, bucket: bucket}
}

func (p *Presigner) PresignedURL(ctx context.Context, key string) (string, error) {
	req, err := p.client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign request, key: %s, %w", key, err)
	}
	return req.URL, nil
}
