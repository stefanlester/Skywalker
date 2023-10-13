package miniofilesystem

import (
	"context"
	"log"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stefanlester/skywalker/filesystems"
)

// Minio is the overall type for the Minio filesystem, containing connection credentials and configuration.
type Minio struct {
	Client *minio.Client
	Bucket string
}

// NewMinio creates a new Minio instance and initializes the Minio client.
func NewMinio(endpoint, key, secret, region, bucket string, useSSL bool) (*Minio, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(key, secret, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	return &Minio{Client: client, Bucket: bucket}, nil
}

// Put transfers a file to the remote filesystem.
func (m *Minio) Put(fileName, folder string) error {
	ctx := context.Background()
	objectName := path.Base(fileName)
	_, err := m.Client.FPutObject(ctx, m.Bucket, path.Join(folder, objectName), fileName, minio.PutObjectOptions{})
	if err != nil {
		log.Printf("Failed to upload %s: %v", objectName, err)
		return err
	}
	return nil
}

// List returns a listing of files in the remote bucket with the given prefix.
func (m *Minio) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	ctx := context.Background()

	objectCh := m.Client.ListObjects(ctx, m.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Failed to list objects: %v", object.Err)
			return listing, object.Err
		}

		if !strings.HasPrefix(object.Key, ".") {
			sizeMB := float64(object.Size) / (1024 * 1024)
			item := filesystems.Listing{
				Etag:         object.ETag,
				LastModified: object.LastModified,
				Key:          object.Key,
				Size:         sizeMB,
			}
			listing = append(listing, item)
		}
	}

	return listing, nil
}

// Delete removes one or more files from the remote filesystem.
func (m *Minio) Delete(itemsToDelete []string) error {
	ctx := context.Background()
	opts := minio.RemoveObjectOptions{GovernanceBypass: true}

	for _, item := range itemsToDelete {
		err := m.Client.RemoveObject(ctx, m.Bucket, item, opts)
		if err != nil {
			log.Printf("Failed to delete %s: %v", item, err)
			return err
		}
	}

	return nil
}

// Get pulls a file from the remote filesystem and saves it locally.
func (m *Minio) Get(destination string, items ...string) error {
	ctx := context.Background()

	for _, item := range items {
		objectName := path.Base(item)
		err := m.Client.FGetObject(ctx, m.Bucket, item, path.Join(destination, objectName), minio.GetObjectOptions{})
		if err != nil {
			log.Printf("Failed to download %s: %v", objectName, err)
			return err
		}
	}

	return nil
}
