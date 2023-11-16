package miniofilesystem

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stefanlester/skywalker/filesystems"
)

const (
	KB = 1024
	MB = KB * KB
)

// Minio is the overall type for the MinIO filesystem.
type Minio struct {
	Endpoint string
	Key      string
	Secret   string
	UseSSL   bool
	Region   string
	Bucket   string
}

// getCredentials generates a Minio client using the credentials stored in the Minio type.
func (m *Minio) getCredentials() (*minio.Client, error) {
	client, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.Key, m.Secret, ""),
		Secure: m.UseSSL,
		Region: m.Region,
		//Bucket:  m.Bucket,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Put transfers a file to the remote file system.
func (m *Minio) Put(ctx context.Context, fileName, folder string) error {
	objectName := path.Base(fileName)
	client, err := m.getCredentials()
	if err != nil {
		return err
	}

	uploadInfo, err := client.FPutObject(ctx, m.Bucket, fmt.Sprintf("%s/%s", folder, objectName), fileName, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	fmt.Println("UploadInfo:", uploadInfo)
	return nil
}

// List returns a listing of all files in the remote bucket with the given prefix.
func (m *Minio) List(ctx context.Context, prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	client, err := m.getCredentials()
	if err != nil {
		return nil, err
	}

	objectCh := client.ListObjects(ctx, m.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return listing, object.Err
		}

		if !strings.HasPrefix(object.Key, ".") {
			sizeMB := float64(object.Size) / MB
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
func (m *Minio) Delete(itemsToDelete []string) bool {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    client, err := m.getCredentials()

	if err != nil {
		fmt.Println("Error getting credentials")
	}

    opts := minio.RemoveObjectOptions{
        GovernanceBypass: true,
    }

    for _, item := range itemsToDelete {
        err := client.RemoveObject(ctx, m.Bucket, item, opts)
        if err != nil {
            fmt.Println(err)
            return false
        }
    }
    return true
}

// Get pulls a file from the remote file system and saves it somewhere on our server.
func (m *Minio) Get(ctx context.Context, destination string, items ...string) error {
	client, err := m.getCredentials()
	if err != nil {
		return err
	}

	for _, item := range items {
		err := client.FGetObject(ctx, m.Bucket, item, fmt.Sprintf("%s/%s", destination, path.Base(item)), minio.GetObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
