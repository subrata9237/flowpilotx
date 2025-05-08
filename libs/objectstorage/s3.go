package objectstorage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3Client struct {
	client *s3.Client
	bucket string
	region string
}

func newS3Client(cfg *Config) (ObjectStorage, error) {
	ctx := context.Background()
	var awsCfg aws.Config
	var err error

	// Load configuration from environment variables or AWS credentials file
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// If credentials are explicitly provided, use them
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				cfg.SessionToken,
			)),
		)
	} else {
		// Otherwise, use the default credential chain
		// This will check in order:
		// 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
		// 2. Shared credentials file (~/.aws/credentials)
		// 3. IAM role for Amazon ECS tasks
		// 4. IAM role for EC2 instances
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, err
	}

	// Create S3 client options
	options := []func(*s3.Options){}
	
	// Add custom endpoint if specified (for MinIO, etc.)
	if cfg.Endpoint != "" {
		options = append(options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			if cfg.ForcePathStyle {
				o.UsePathStyle = true
			}
		})
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, options...)

	return &s3Client{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}

// Helper function to get client and bucket for an operation
func (c *s3Client) getClientAndBucket(ctx context.Context, commonOpts objectstorage.CommonOptions) (*s3.Client, string, error) {
	bucket := c.bucket
	client := c.client

	if commonOpts.Bucket != "" {
		bucket = commonOpts.Bucket
	}

	if commonOpts.Region != "" && commonOpts.Region != c.region {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(commonOpts.Region),
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create client for region %s: %w", commonOpts.Region, err)
		}
		client = s3.NewFromConfig(cfg)
	}

	return client, bucket, nil
}

func (c *s3Client) Upload(ctx context.Context, key string, reader io.Reader, opts *UploadOptions) (*VersionInfo, error) {
	if opts == nil {
		opts = &UploadOptions{}
	}

	// Determine which bucket to use
	bucket := c.bucket
	if opts.CommonOptions.Bucket != "" {
		bucket = opts.CommonOptions.Bucket
	}

	// If a different region is specified, create a new client for that region
	client := c.client
	if opts.CommonOptions.Region != "" && opts.CommonOptions.Region != c.region {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(opts.CommonOptions.Region),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create client for region %s: %w", opts.CommonOptions.Region, err)
		}
		client = s3.NewFromConfig(cfg)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}
	if opts.ContentEncoding != "" {
		input.ContentEncoding = aws.String(opts.ContentEncoding)
	}
	if opts.CacheControl != "" {
		input.CacheControl = aws.String(opts.CacheControl)
	}
	if !opts.Expires.IsZero() {
		input.Expires = aws.Time(opts.Expires)
	}
	if opts.StorageClass != "" {
		input.StorageClass = types.StorageClass(opts.StorageClass)
	}

	// Handle ACL options
	if opts.ACL != nil {
		if opts.ACL.PredefinedACL != "" {
			// Use predefined ACL
			input.ACL = types.ObjectCannedACL(opts.ACL.PredefinedACL)
		} else if len(opts.ACL.Grants) > 0 {
			// Use custom grants
			grants := make([]types.Grant, 0, len(opts.ACL.Grants))
			for _, grant := range opts.ACL.Grants {
				grantee := types.Grantee{
					Type: types.Type(grant.GranteeType),
				}

				switch grant.GranteeType {
				case objectstorage.GranteeTypeCanonicalUser:
					grantee.ID = aws.String(grant.Grantee)
				case objectstorage.GranteeTypeAmazonCustomerByEmail:
					grantee.EmailAddress = aws.String(grant.Grantee)
				case objectstorage.GranteeTypeGroup:
					grantee.URI = aws.String(grant.Grantee)
				}

				grants = append(grants, types.Grant{
					Grantee:    &grantee,
					Permission: types.Permission(grant.Permission),
				})
			}

			input.GrantFullControl = nil // Clear any canned ACL
			input.GrantRead = nil
			input.GrantReadACP = nil
			input.GrantWrite = nil
			input.GrantWriteACP = nil
			input.ACL = ""

			input.Grants = grants
		}
	}

	// Initialize metadata map if nil
	if input.Metadata == nil {
		input.Metadata = make(map[string]string)
	}
	input.Metadata["created-at"] = time.Now().UTC().Format(time.RFC3339)

	// Add any custom metadata
	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			input.Metadata[k] = v
		}
	}

	// Add region information to metadata if using non-default region
	if opts.CommonOptions.Region != "" {
		input.Metadata["upload-region"] = opts.CommonOptions.Region
	}

	result, err := client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
	}

	return &VersionInfo{
		VersionID:    aws.ToString(result.VersionId),
		LastModified: time.Now().UTC(),
		IsLatest:     true,
		Size:         0, // Size not available in PutObjectOutput
	}, nil
}

func (c *s3Client) Download(ctx context.Context, key string, opts *objectstorage.DownloadOptions) (io.ReadCloser, error) {
	if opts == nil {
		opts = &objectstorage.DownloadOptions{}
	}

	client, bucket, err := c.getClientAndBucket(ctx, opts.CommonOptions)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if opts.VersionID != "" {
		input.VersionId = aws.String(opts.VersionID)
	}

	output, err := client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}

	return output.Body, nil
}

func (c *s3Client) Delete(ctx context.Context, key string, opts *objectstorage.DeleteOptions) error {
	if opts == nil {
		opts = &objectstorage.DeleteOptions{}
	}

	client, bucket, err := c.getClientAndBucket(ctx, opts.CommonOptions)
	if err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if opts.VersionID != "" {
		input.VersionId = aws.String(opts.VersionID)
	}

	_, err = client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func (c *s3Client) List(ctx context.Context, opts *objectstorage.ListOptions) ([]objectstorage.ObjectInfo, error) {
	if opts == nil {
		opts = &objectstorage.ListOptions{}
	}

	client, bucket, err := c.getClientAndBucket(ctx, opts.CommonOptions)
	if err != nil {
		return nil, err
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if opts.Prefix != "" {
		input.Prefix = aws.String(opts.Prefix)
	}
	if opts.Delimiter != "" {
		input.Delimiter = aws.String(opts.Delimiter)
	}
	if opts.MaxKeys > 0 {
		input.MaxKeys = aws.Int32(int32(opts.MaxKeys))
	}

	output, err := client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	objects := make([]objectstorage.ObjectInfo, 0, len(output.Contents))
	for _, obj := range output.Contents {
		objects = append(objects, objectstorage.ObjectInfo{
			Key:          *obj.Key,
			Size:         obj.Size,
			LastModified: *obj.LastModified,
			ETag:         *obj.ETag,
		})
	}

	return objects, nil
}

func (c *s3Client) GetInfo(ctx context.Context, key string, opts *GetInfoOptions) (*ObjectInfo, error) {
	client, bucket, err := c.getClientAndBucket(ctx, &opts.CommonOptions)
	if err != nil {
		return nil, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if opts != nil && opts.VersionID != "" {
		input.VersionId = aws.String(opts.VersionID)
	}

	output, err := client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &ObjectInfo{
		Key:          key,
		Size:         output.ContentLength,
		LastModified: *output.LastModified,
		CreatedAt:    getCreationTime(output.Metadata),
		ContentType:  aws.ToString(output.ContentType),
		ETag:         aws.ToString(output.ETag),
		Metadata:     output.Metadata,
		StorageClass: string(output.StorageClass),
		VersionID:    aws.ToString(output.VersionId),
	}, nil
}

func (c *s3Client) GeneratePresignedURL(ctx context.Context, key string, expires time.Duration, opts *CommonOptions) (string, error) {
	client, bucket, err := c.getClientAndBucket(ctx, opts)
	if err != nil {
		return "", err
	}

	presignClient := s3.NewPresignClient(client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	presignedURL, err := presignClient.PresignGetObject(ctx, input,
		s3.WithPresignExpires(expires),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.URL, nil
}

func (c *s3Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if the error is a "not found" error
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *s3Client) GetModificationTime(ctx context.Context, key string) (time.Time, error) {
	output, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return time.Time{}, err
	}
	return *output.LastModified, nil
}

func (c *s3Client) GetCreationTime(ctx context.Context, key string) (time.Time, error) {
	output, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return time.Time{}, err
	}

	// Try to get creation time from metadata
	if t := getCreationTime(output.Metadata); !t.IsZero() {
		return t, nil
	}

	// Fall back to last modified time if creation time is not available
	return *output.LastModified, nil
}

// Helper function to extract creation time from metadata
func getCreationTime(metadata map[string]string) time.Time {
	if createdAt, ok := metadata["created-at"]; ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (c *s3Client) Close() error {
	// AWS SDK v2 doesn't require explicit cleanup
	return nil
}

// GetVersions gets all versions of an object
func (c *s3Client) GetVersions(ctx context.Context, key string, opts *objectstorage.CommonOptions) ([]objectstorage.VersionInfo, error) {
	if opts == nil {
		opts = &objectstorage.CommonOptions{}
	}

	client, bucket, err := c.getClientAndBucket(ctx, *opts)
	if err != nil {
		return nil, err
	}

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	}

	result, err := client.ListObjectVersions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list object versions: %w", err)
	}

	versions := make([]objectstorage.VersionInfo, 0, len(result.Versions))
	for _, version := range result.Versions {
		versions = append(versions, objectstorage.VersionInfo{
			VersionID:    aws.ToString(version.VersionId),
			LastModified: *version.LastModified,
			IsLatest:     *version.IsLatest,
			Size:         version.Size,
		})
	}

	return versions, nil
}

// GetLatestVersion gets the latest version of an object
func (c *s3Client) GetLatestVersion(ctx context.Context, key string, opts *objectstorage.CommonOptions) (*objectstorage.VersionInfo, error) {
	if opts == nil {
		opts = &objectstorage.CommonOptions{}
	}

	client, bucket, err := c.getClientAndBucket(ctx, *opts)
	if err != nil {
		return nil, err
	}

	input := &s3.ListObjectVersionsInput{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(key),
		MaxKeys: aws.Int32(1),
	}

	result, err := client.ListObjectVersions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	if len(result.Versions) == 0 {
		return nil, fmt.Errorf("no versions found for key: %s", key)
	}

	latestVersion := result.Versions[0]
	return &objectstorage.VersionInfo{
		VersionID:    aws.ToString(latestVersion.VersionId),
		LastModified: *latestVersion.LastModified,
		IsLatest:     true,
		Size:         latestVersion.Size,
	}, nil
} 