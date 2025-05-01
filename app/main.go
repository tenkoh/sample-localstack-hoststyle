package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tenkoh/sample-localstack-hoststyle/presigner"
)

const topTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LocalStack example: S3 virtual host style URL</title>
</head>
<body>
  {{ range $index, $content := . }}
	<h2>{{ $content.Title }}</h2>
	<p>Download from: {{ $content.Host }}</p>
	<a href="{{ $content.PresignedUrl }}">Download</a>
  {{ end }}
</body>
</html>
`

type content struct {
	Title        string
	Host         string
	PresignedUrl string
}

func parseContent(title, urlString string) (content, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return content{}, fmt.Errorf("failed to parse url: %s, %w", urlString, err)
	}
	return content{
		Title:        title,
		Host:         u.Hostname(),
		PresignedUrl: urlString,
	}, nil
}

func mainHandler(s3Client *s3.PresignClient, lambdaClient *lambda.Client, bucket, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		s3Content, err := getPresignInApp(r.Context(), s3Client, bucket, key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get presign in app: %v", err), http.StatusInternalServerError)
			return
		}

		lambdaContent, err := getPresignInLambda(r.Context(), lambdaClient, bucket, key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get presign in lambda: %v", err), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.New("top").Parse(topTemplate))
		if err := tmpl.Execute(w, []content{s3Content, lambdaContent}); err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func getPresignInApp(ctx context.Context, client *s3.PresignClient, bucket string, key string) (content, error) {
	p, err := presigner.NewPresigner(client, bucket).PresignedURL(ctx, key)
	if err != nil {
		return content{}, fmt.Errorf("failed to presign URL: %w", err)
	}
	return parseContent("Created by SDK in app", p)
}

func getPresignInLambda(ctx context.Context, client *lambda.Client, bucket string, key string) (content, error) {
	payload, err := json.Marshal(map[string]string{"key": key})
	if err != nil {
		return content{}, fmt.Errorf("failed to marshal payload: %w", err)
	}
	req := &lambda.InvokeInput{
		FunctionName: aws.String("presign-function"),
		Payload:      payload,
	}
	resp, err := client.Invoke(ctx, req)
	if err != nil {
		return content{}, fmt.Errorf("failed to invoke lambda: %w", err)
	}
	if resp.StatusCode != 200 {
		return content{}, fmt.Errorf("lambda returned non-200 status code: %d", resp.StatusCode)
	}
	var presignResponse struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(resp.Payload, &presignResponse); err != nil {
		return content{}, fmt.Errorf("failed to unmarshal lambda response: %w", err)
	}
	return parseContent("Created by SDK in Lambda", presignResponse.URL)
}

func main() {
	ctx := context.Background()
	bucket := os.Getenv("BUCKET_NAME")
	objectKey := "test.txt"

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// アプリケーションの中で署名付きURLを生成
	s3Client := s3.NewPresignClient(s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT_URL"))
	}))

	// Lambdaを呼び出して署名付きURLを生成
	lambdaClient := lambda.NewFromConfig(cfg, func(o *lambda.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("OTHER_ENDPOINT_URL"))
	})

	http.HandleFunc("/", mainHandler(s3Client, lambdaClient, bucket, objectKey))

	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
