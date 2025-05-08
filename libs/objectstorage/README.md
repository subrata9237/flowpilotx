# Object Storage Library

A unified object storage interface supporting multiple cloud providers (AWS S3, Google Cloud Storage, etc.) with a consistent API.

## Features

- Unified interface for multiple storage providers
- Automatic retries and error handling
- Streaming upload/download support
- Concurrent operations
- Progress tracking
- Pre-signed URL generation
- Metadata management

## Supported Providers

- Amazon S3
- Google Cloud Storage
- MinIO
- Local filesystem (for development)

## Installation

```bash
go get github.com/flowpilotx/libs/objectstorage
```

## Configuration

```go
type Config struct {
    Provider     string // "s3", "gcs", "minio", "local"
    Endpoint     string // Optional custom endpoint
    Region       string
    Bucket       string
    Credentials  CredentialsConfig
    MaxRetries   int
    Timeout      time.Duration
}

type CredentialsConfig struct {
    AccessKey    string
    SecretKey    string
    TokenPath    string // For service account credentials
}
```

## Usage Examples

### Basic Operations

```go
package main

import (
    "context"
    "log"
    
    "github.com/flowpilotx/libs/objectstorage"
)

func main() {
    // Initialize client
    config := objectstorage.Config{
        Provider: "s3",
        Region:   "us-west-2",
        Bucket:   "my-bucket",
        Credentials: objectstorage.CredentialsConfig{
            AccessKey: "access-key",
            SecretKey: "secret-key",
        },
    }
    
    client, err := objectstorage.NewClient(config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    ctx := context.Background()
    
    // Upload file
    err = client.Upload(ctx, "local/path/file.txt", "remote/path/file.txt")
    if err != nil {
        log.Fatalf("Upload failed: %v", err)
    }
    
    // Download file
    err = client.Download(ctx, "remote/path/file.txt", "local/path/file.txt")
    if err != nil {
        log.Fatalf("Download failed: %v", err)
    }
    
    // Delete file
    err = client.Delete(ctx, "remote/path/file.txt")
    if err != nil {
        log.Fatalf("Delete failed: %v", err)
    }
}
```

### Streaming Operations

```go
package main

import (
    "context"
    "io"
    "log"
    "os"
    
    "github.com/flowpilotx/libs/objectstorage"
)

func main() {
    client, err := objectstorage.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Upload stream
    file, _ := os.Open("large-file.dat")
    defer file.Close()
    
    err = client.UploadStream(ctx, file, "remote/path/large-file.dat", &objectstorage.UploadOptions{
        ContentType: "application/octet-stream",
        Metadata: map[string]string{
            "source": "example",
        },
        ProgressFunc: func(bytesWritten int64) {
            log.Printf("Uploaded %d bytes", bytesWritten)
        },
    })
    
    // Download stream
    reader, err := client.DownloadStream(ctx, "remote/path/large-file.dat")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()
    
    outFile, _ := os.Create("downloaded-file.dat")
    defer outFile.Close()
    
    _, err = io.Copy(outFile, reader)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Pre-signed URLs

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/flowpilotx/libs/objectstorage"
)

func main() {
    client, err := objectstorage.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Generate pre-signed URL for download
    url, err := client.GetPresignedURL(ctx, "remote/path/file.txt", time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Download URL: %s", url)
    
    // Generate pre-signed URL for upload
    uploadURL, err := client.GetPresignedUploadURL(ctx, "remote/path/new-file.txt", time.Hour, &objectstorage.UploadOptions{
        ContentType: "text/plain",
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Upload URL: %s", uploadURL)
}
```

## Error Handling

The library provides detailed error types for better error handling:

```go
switch err := err.(type) {
case *objectstorage.NotFoundError:
    log.Printf("Object not found: %s", err.Path)
case *objectstorage.AccessDeniedError:
    log.Printf("Access denied: %s", err.Message)
case *objectstorage.NetworkError:
    log.Printf("Network error: %s", err.Message)
default:
    log.Printf("Unknown error: %v", err)
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Running Examples

```bash
# Set up environment variables
export STORAGE_PROVIDER=s3
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-west-2
export BUCKET_NAME=your-bucket

# Run example
go run examples/main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit a pull request

## License

This library is licensed under the MIT License. 