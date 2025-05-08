package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/flowpilotx/libs/objectstorage"
)

func main() {
	// Example 1: Using explicit credentials
	explicitConfig := &objectstorage.Config{
		Provider:        objectstorage.ProviderS3,
		Region:         "us-west-2",
		Bucket:         "my-test-bucket",
		AccessKeyID:     "your-access-key",
		SecretAccessKey: "your-secret-key",
		Endpoint:       "http://localhost:9000", // For MinIO
		ForcePathStyle: true,                    // For MinIO
	}

	// Example 2: Using environment variables or AWS credentials
	// Make sure to set these environment variables:
	// export AWS_ACCESS_KEY_ID=your-access-key
	// export AWS_SECRET_ACCESS_KEY=your-secret-key
	// export AWS_REGION=us-west-2
	envConfig := &objectstorage.Config{
		Provider: objectstorage.ProviderS3,
		Region:   os.Getenv("AWS_REGION"), // Will use AWS_REGION environment variable
		Bucket:   "my-test-bucket",
	}

	// Choose which config to use
	cfg := envConfig // or explicitConfig

	// Create storage client
	storage, err := objectstorage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}
	defer storage.Close()

	// Example usage
	err = uploadAndDownloadExample(storage)
	if err != nil {
		log.Fatalf("Example failed: %v", err)
	}

	// Run upload examples
	if err := uploadExamples(storage); err != nil {
		log.Fatalf("Upload examples failed: %v", err)
	}

	// Example usage with different buckets and regions
	err = demonstrateOperations(storage)
	if err != nil {
		log.Fatalf("Operations failed: %v", err)
	}

	// Demonstrate ACL operations
	if err := demonstrateACLOperations(storage); err != nil {
		log.Fatalf("ACL operations failed: %v", err)
	}
}

func uploadAndDownloadExample(storage objectstorage.ObjectStorage) error {
	ctx := context.Background()

	// Check if file exists before upload
	exists, err := storage.Exists(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}
	if exists {
		fmt.Println("File already exists")
	} else {
		fmt.Println("File does not exist, proceeding with upload")
	}

	// Upload a file with storage class
	content := strings.NewReader("Hello, World!")
	err = storage.Upload(ctx, "test.txt", content, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		Metadata: map[string]string{
			"created-by": "example",
		},
		StorageClass: "STANDARD", // or "STANDARD_IA", "GLACIER", etc.
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	fmt.Println("File uploaded successfully")

	// Get file info with timestamps
	info, err := storage.GetInfo(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fmt.Printf("File info:\n")
	fmt.Printf("  Key: %s\n", info.Key)
	fmt.Printf("  Size: %d bytes\n", info.Size)
	fmt.Printf("  Created: %v\n", info.CreatedAt)
	fmt.Printf("  Modified: %v\n", info.LastModified)
	fmt.Printf("  Storage Class: %s\n", info.StorageClass)
	fmt.Printf("  Version ID: %s\n", info.VersionID)
	fmt.Printf("  Metadata: %v\n", info.Metadata)

	// Get specific timestamps
	modTime, err := storage.GetModificationTime(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to get modification time: %w", err)
	}
	fmt.Printf("Modification time: %v\n", modTime)

	createTime, err := storage.GetCreationTime(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to get creation time: %w", err)
	}
	fmt.Printf("Creation time: %v\n", createTime)

	// Generate presigned URL
	url, err := storage.GeneratePresignedURL(ctx, "test.txt", 1*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	fmt.Printf("Presigned URL: %s\n", url)

	// List files
	files, err := storage.List(ctx, &objectstorage.ListOptions{
		Prefix: "test",
	})
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	fmt.Println("\nFiles found:")
	for _, f := range files {
		fmt.Printf("  - %s (created: %v, modified: %v)\n", 
			f.Key, f.CreatedAt, f.LastModified)
	}

	// Upload a file
	content = strings.NewReader("Hello, World!")
	uploadOpts := &objectstorage.UploadOptions{
		ContentType: "text/plain",
		Metadata: map[string]string{
			"custom-key": "custom-value",
		},
	}

	versionInfo, err := storage.Upload(context.Background(), "test.txt", content, uploadOpts)
	if err != nil {
		log.Fatalf("Failed to upload: %v", err)
	}
	fmt.Printf("Uploaded file version: %s\n", versionInfo.VersionID)

	// Upload another version of the same file
	newContent := strings.NewReader("Hello, Updated World!")
	newVersionInfo, err := storage.Upload(context.Background(), "test.txt", newContent, uploadOpts)
	if err != nil {
		log.Fatalf("Failed to upload new version: %v", err)
	}
	fmt.Printf("Uploaded new version: %s\n", newVersionInfo.VersionID)

	// Get all versions of the file
	versions, err := storage.GetVersions(context.Background(), "test.txt")
	if err != nil {
		log.Fatalf("Failed to get versions: %v", err)
	}

	fmt.Println("\nFile versions:")
	for _, v := range versions {
		fmt.Printf("- Version ID: %s, Last Modified: %s, Is Latest: %v, Size: %d bytes\n",
			v.VersionID, v.LastModified.Format(time.RFC3339), v.IsLatest, v.Size)
	}

	// Get the latest version
	latestVersion, err := storage.GetLatestVersion(context.Background(), "test.txt")
	if err != nil {
		log.Fatalf("Failed to get latest version: %v", err)
	}
	fmt.Printf("\nLatest version details:\n")
	fmt.Printf("Version ID: %s\n", latestVersion.VersionID)
	fmt.Printf("Last Modified: %s\n", latestVersion.LastModified.Format(time.RFC3339))
	fmt.Printf("Size: %d bytes\n", latestVersion.Size)

	// Download file
	reader, err := storage.Download(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer reader.Close()

	// Delete file
	err = storage.Delete(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	fmt.Println("\nFile deleted successfully")

	// Verify file no longer exists
	exists, err = storage.Exists(ctx, "test.txt")
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		fmt.Println("File deletion confirmed")
	}

	return nil
}

func uploadExamples(storage objectstorage.ObjectStorage) error {
	ctx := context.Background()

	// Example 1: Upload to default bucket and region
	content1 := strings.NewReader("Hello from default bucket!")
	versionInfo1, err := storage.Upload(ctx, "test1.txt", content1, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		Metadata: map[string]string{
			"source": "example1",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to default bucket: %w", err)
	}
	fmt.Printf("Uploaded to default bucket, version: %s\n", versionInfo1.VersionID)

	// Example 2: Upload to a different bucket in the same region
	content2 := strings.NewReader("Hello from another bucket!")
	versionInfo2, err := storage.Upload(ctx, "test2.txt", content2, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		Bucket:      "another-bucket",
		Metadata: map[string]string{
			"source": "example2",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to another bucket: %w", err)
	}
	fmt.Printf("Uploaded to bucket 'another-bucket', version: %s\n", versionInfo2.VersionID)

	// Example 3: Upload to a different bucket and region
	content3 := strings.NewReader("Hello from another region!")
	versionInfo3, err := storage.Upload(ctx, "test3.txt", content3, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		Bucket:      "cross-region-bucket",
		Region:      "eu-west-1", // Different region
		Metadata: map[string]string{
			"source": "example3",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to another region: %w", err)
	}
	fmt.Printf("Uploaded to bucket 'cross-region-bucket' in region 'eu-west-1', version: %s\n", versionInfo3.VersionID)

	return nil
}

func demonstrateOperations(storage objectstorage.ObjectStorage) error {
	ctx := context.Background()

	// Upload to default bucket
	content := strings.NewReader("Hello, World!")
	versionInfo, err := storage.Upload(ctx, "test.txt", content, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		Metadata: map[string]string{
			"source": "example",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to default bucket: %w", err)
	}
	fmt.Printf("Uploaded to default bucket, version: %s\n", versionInfo.VersionID)

	// Upload to different bucket and region
	crossRegionContent := strings.NewReader("Hello from another region!")
	crossRegionVersion, err := storage.Upload(ctx, "test.txt", crossRegionContent, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		CommonOptions: objectstorage.CommonOptions{
			Bucket: "cross-region-bucket",
			Region: "eu-west-1",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload to cross-region bucket: %w", err)
	}
	fmt.Printf("Uploaded to cross-region bucket, version: %s\n", crossRegionVersion.VersionID)

	// List objects in different bucket
	listOpts := &objectstorage.ListOptions{
		CommonOptions: objectstorage.CommonOptions{
			Bucket: "another-bucket",
		},
		Prefix: "test",
	}
	objects, err := storage.List(ctx, listOpts)
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}
	fmt.Printf("\nObjects in another-bucket:\n")
	for _, obj := range objects {
		fmt.Printf("- %s (size: %d bytes)\n", obj.Key, obj.Size)
	}

	// Get object info from specific version
	infoOpts := &objectstorage.GetInfoOptions{
		CommonOptions: objectstorage.CommonOptions{
			Bucket: "cross-region-bucket",
			Region: "eu-west-1",
		},
		VersionID: crossRegionVersion.VersionID,
	}
	info, err := storage.GetInfo(ctx, "test.txt", infoOpts)
	if err != nil {
		return fmt.Errorf("failed to get object info: %w", err)
	}
	fmt.Printf("\nObject info from cross-region bucket:\n")
	fmt.Printf("Key: %s\nSize: %d\nVersion: %s\n", info.Key, info.Size, info.VersionID)

	// Download from specific version
	downloadOpts := &objectstorage.DownloadOptions{
		CommonOptions: objectstorage.CommonOptions{
			Bucket: "cross-region-bucket",
			Region: "eu-west-1",
		},
		VersionID: crossRegionVersion.VersionID,
	}
	reader, err := storage.Download(ctx, "test.txt", downloadOpts)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer reader.Close()

	// Get versions from a specific bucket
	versionsOpts := &objectstorage.CommonOptions{
		Bucket: "cross-region-bucket",
		Region: "eu-west-1",
	}
	versions, err := storage.GetVersions(ctx, "test.txt", versionsOpts)
	if err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}
	fmt.Printf("\nVersions in cross-region bucket:\n")
	for _, v := range versions {
		fmt.Printf("- Version: %s, Last Modified: %s\n", v.VersionID, v.LastModified)
	}

	// Generate presigned URL for cross-region bucket
	url, err := storage.GeneratePresignedURL(ctx, "test.txt", time.Hour, versionsOpts)
	if err != nil {
		return fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	fmt.Printf("\nPresigned URL for cross-region bucket: %s\n", url)

	// Delete specific version from cross-region bucket
	deleteOpts := &objectstorage.DeleteOptions{
		CommonOptions: objectstorage.CommonOptions{
			Bucket: "cross-region-bucket",
			Region: "eu-west-1",
		},
		VersionID: crossRegionVersion.VersionID,
	}
	err = storage.Delete(ctx, "test.txt", deleteOpts)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	fmt.Printf("\nDeleted version %s from cross-region bucket\n", crossRegionVersion.VersionID)

	return nil
}

func demonstrateACLOperations(storage objectstorage.ObjectStorage) error {
	ctx := context.Background()

	// Example 1: Upload with predefined ACL (public-read)
	content1 := strings.NewReader("This is a public file")
	_, err := storage.Upload(ctx, "public.txt", content1, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		ACL: &objectstorage.ACLOptions{
			PredefinedACL: objectstorage.ACLPublicRead,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload public file: %w", err)
	}
	fmt.Println("Uploaded file with public-read ACL")

	// Example 2: Upload with custom grants
	content2 := strings.NewReader("This file has custom permissions")
	customGrants := []objectstorage.Grant{
		{
			// Grant read access to a specific AWS account by email
			Grantee:     "user@example.com",
			GranteeType: objectstorage.GranteeTypeAmazonCustomerByEmail,
			Permission:  objectstorage.PermissionRead,
		},
		{
			// Grant full control to a specific canonical user
			Grantee:     "79a59df900b949e55d96a1e698fbacedfd6e09d98eacf8f8d5218e7cd47ef2be",
			GranteeType: objectstorage.GranteeTypeCanonicalUser,
			Permission:  objectstorage.PermissionFullControl,
		},
		{
			// Grant read access to all authenticated AWS users
			Grantee:     "http://acs.amazonaws.com/groups/global/AuthenticatedUsers",
			GranteeType: objectstorage.GranteeTypeGroup,
			Permission:  objectstorage.PermissionRead,
		},
	}

	_, err = storage.Upload(ctx, "custom-acl.txt", content2, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		ACL: &objectstorage.ACLOptions{
			Grants: customGrants,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload file with custom ACL: %w", err)
	}
	fmt.Println("Uploaded file with custom ACL grants")

	// Example 3: Upload with bucket owner full control
	content3 := strings.NewReader("Bucket owner has full control")
	_, err = storage.Upload(ctx, "bucket-owner.txt", content3, &objectstorage.UploadOptions{
		ContentType: "text/plain",
		ACL: &objectstorage.ACLOptions{
			PredefinedACL: objectstorage.ACLBucketOwnerFullControl,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload file with bucket owner control: %w", err)
	}
	fmt.Println("Uploaded file with bucket-owner-full-control ACL")

	return nil
} 