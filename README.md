# sample-localstack-hoststyle
A sample architecture using LocalStack as its local development environment which never uses "path style" S3 URL.

## Prerequisites
- Docker (including Docker Compose)
- Go
- make (NOTE: alternatively, you can do the same commands in the Makefile)

## Overview

This application demonstrates how to generate and use pre-signed URLs for S3 objects in a local development environment using LocalStack. The URLs are generated both from the application container and Lambda function running in LocalStack.

```mermaid
sequenceDiagram
    participant Browser
    participant App as Application Container
    participant S3 as LocalStack S3
    participant Lambda as Lambda Container

    Browser->>App: Access top page
    App->>S3: Generate pre-signed URL (SDK)
    S3-->>App: Return URL 1
    App->>Lambda: Request pre-signed URL
    Lambda->>S3: Generate pre-signed URL (SDK)
    S3-->>Lambda: Return URL
    Lambda-->>App: Return URL 2
    App-->>Browser: Return page with both URLs
    
    alt Download using URL 1
        Browser->>S3: Access URL 1
        S3-->>Browser: Return file
    else Download using URL 2
        Browser->>S3: Access URL 2
        S3-->>Browser: Return file
    end
```

## Author
tenkoh

## License
MIT
