package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const topTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LocalStack example: S3 virtual host style URL</title>
</head>
<body>
  <h2>{{ .AnotherContainer.Title }}</h2>
  <p>Download from: {{ .AnotherContainer.Host }}</p>
  <a href="{{ .AnotherContainer.PresignedUrl }}">Download</a>
</body>
</html>
`

type topContent struct {
	AnotherContainer content
}

type content struct {
	Title        string
	Host         string
	PresignedUrl string
}

type presigner struct {
	client *s3.PresignClient
	bucket string
}

func NewPresigner(ctx context.Context, bucket string) (*presigner, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// o.EndpointResolverV2 = &resolverV2{}
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT_URL"))
	})
	return &presigner{
		client: s3.NewPresignClient(client),
		bucket: bucket,
	}, nil
}

func (p *presigner) PresignedURL(ctx context.Context, key string) (string, error) {
	req, err := p.client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign request, key: %s, %w", key, err)
	}
	return req.URL, nil
}

func mustParseUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse url: %s, %v", s, err))
	}
	return u
}

func mainHandler(p *presigner, targetKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pu, err := p.PresignedURL(r.Context(), targetKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get object. Key: %s. Detail: %v", targetKey, err), http.StatusNotFound)
			return
		}
		u := mustParseUrl(pu)
		host := u.Hostname()

		// view
		w.Header().Set("Content-Type", "text/html")
		content := topContent{
			AnotherContainer: content{
				Title:        "Created by SDK in another container",
				Host:         host,
				PresignedUrl: pu,
			},
		}
		tmpl := template.Must(template.New("top").Parse(topTemplate))
		if tmpl.Execute(w, content) != nil {
			http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	ctx := context.Background()
	bucket := os.Getenv("BUCKET_NAME")
	presigner, err := NewPresigner(ctx, bucket)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", mainHandler(presigner, "test.txt"))

	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
