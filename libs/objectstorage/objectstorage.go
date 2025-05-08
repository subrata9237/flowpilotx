package objectstorage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Provider represents the type of storage provider
type Provider string

const (
	// ProviderS3 represents Amazon S3 storage
	ProviderS3 Provider = "s3"
	// ProviderGCS represents Google Cloud Storage
	ProviderGCS Provider = "gcs"
	// ProviderAzure represents Azure Blob Storage
	ProviderAzure Provider = "azure"
)

// ObjectInfo contains metadata about a stored object
type ObjectInfo struct {
	Key           string
	Size          int64
	LastModified  time.Time
	CreatedAt     time.Time    // Creation time if available
	ContentType   string
	ETag          string
	Metadata      map[string]string
	StorageClass  string       // Storage class (e.g., STANDARD, GLACIER)
	VersionID     string       // Version ID if versioning is enabled
}

// CommonOptions contains common options for all operations
type CommonOptions struct {
	Bucket string // Optional bucket name, if not provided uses default bucket
	Region string // Optional region, if not provided uses default region
}

// ListOptions contains options for listing objects
type ListOptions struct {
	CommonOptions
	Prefix    string
	Delimiter string
	MaxKeys   int64
	Marker    string
}

// ObjectACL represents predefined ACL values
type ObjectACL string

const (
	// Predefined ACLs
	ACLPrivate                ObjectACL = "private"
	ACLPublicRead            ObjectACL = "public-read"
	ACLPublicReadWrite       ObjectACL = "public-read-write"
	ACLAuthenticatedRead     ObjectACL = "authenticated-read"
	ACLBucketOwnerRead       ObjectACL = "bucket-owner-read"
	ACLBucketOwnerFullControl ObjectACL = "bucket-owner-full-control"
)

// Permission represents the type of permission granted
type Permission string

const (
	PermissionRead         Permission = "READ"
	PermissionWrite        Permission = "WRITE"
	PermissionReadACP      Permission = "READ_ACP"
	PermissionWriteACP     Permission = "WRITE_ACP"
	PermissionFullControl  Permission = "FULL_CONTROL"
)

// GranteeType represents the type of grantee
type GranteeType string

const (
	GranteeTypeCanonicalUser GranteeType = "CanonicalUser"
	GranteeTypeAmazonCustomerByEmail GranteeType = "AmazonCustomerByEmail"
	GranteeTypeGroup GranteeType = "Group"
)

// Grant represents an ACL grant
type Grant struct {
	Grantee     string      // Email address, canonical ID, or group URI
	GranteeType GranteeType // Type of grantee
	Permission  Permission  // Permission to grant
}

// ACLOptions represents options for setting object ACL
type ACLOptions struct {
	PredefinedACL ObjectACL // Predefined ACL to use
	Grants        []Grant   // Custom grants (if PredefinedACL is not set)
}

// UploadOptions contains options for uploading objects
type UploadOptions struct {
	CommonOptions
	ContentType     string
	Metadata        map[string]string
	ContentEncoding string
	CacheControl    string
	Expires         time.Time
	StorageClass    string    // Storage class for the object
	VersionID       string    // Optional version ID for updating specific version
	ACL            *ACLOptions // Access control options
}

// DownloadOptions contains options for downloading objects
type DownloadOptions struct {
	CommonOptions
	VersionID string // Optional specific version to download
}

// DeleteOptions contains options for deleting objects
type DeleteOptions struct {
	CommonOptions
	VersionID string // Optional specific version to delete
}

// GetInfoOptions contains options for getting object info
type GetInfoOptions struct {
	CommonOptions
	VersionID string // Optional specific version to get info about
}

// VersionInfo contains information about an object version
type VersionInfo struct {
	VersionID    string
	LastModified time.Time
	IsLatest     bool
	Size         int64
}

// ObjectStorage defines the interface for object storage operations
type ObjectStorage interface {
	// Upload uploads an object to storage
	Upload(ctx context.Context, key string, reader io.Reader, opts *UploadOptions) (*VersionInfo, error)

	// Download downloads an object from storage
	Download(ctx context.Context, key string, opts *DownloadOptions) (io.ReadCloser, error)

	// Delete removes an object from storage
	Delete(ctx context.Context, key string, opts *DeleteOptions) error

	// List lists objects in storage
	List(ctx context.Context, opts *ListOptions) ([]ObjectInfo, error)

	// GetInfo gets information about an object
	GetInfo(ctx context.Context, key string, opts *GetInfoOptions) (*ObjectInfo, error)

	// GeneratePresignedURL generates a presigned URL for temporary access
	GeneratePresignedURL(ctx context.Context, key string, expires time.Duration, opts *CommonOptions) (string, error)

	// GetVersions gets all versions of an object
	GetVersions(ctx context.Context, key string, opts *CommonOptions) ([]VersionInfo, error)

	// GetLatestVersion gets the latest version of an object
	GetLatestVersion(ctx context.Context, key string, opts *CommonOptions) (*VersionInfo, error)

	// Close closes the storage client
	Close() error
}

// Config contains configuration for object storage
type Config struct {
	Provider Provider
	Region   string
	Bucket   string

	// AWS specific configuration
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Endpoint        string // For custom endpoints (e.g., MinIO)
	ForcePathStyle  bool   // For S3-compatible services

	// Common configuration
	Timeout         time.Duration
	MaxRetries      int
	RetryInterval   time.Duration
	SSLEnabled      bool
	MaxConnections  int
}

// New creates a new ObjectStorage instance based on the provider
func New(cfg *Config) (ObjectStorage, error) {
	switch cfg.Provider {
	case ProviderS3:
		return newS3Client(cfg)
	case ProviderGCS:
		return nil, fmt.Errorf("GCS provider not implemented yet")
	case ProviderAzure:
		return nil, fmt.Errorf("Azure provider not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
} 