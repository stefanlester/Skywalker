package miniofilesystem

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stefanlester/skywalker/filesystems"
)

// Minio is the overall type for the minio filesystem, and contains
// the connection credentials, endpoint, and the bucket to use
type Minio struct {
	Endpoint string
	Key      string
	Secret   string
	UseSSL   bool
	Region   string
	Bucket   string

	clientOnce sync.Once
	client     *minio.Client
}

// getCredentials returns a minio client built from the credentials stored in
// the Minio type. The client is created once per Minio instance and reused;
// minio-go clients are safe for concurrent use.
func (m *Minio) getCredentials() *minio.Client {
	m.clientOnce.Do(func() {
		client, err := minio.New(m.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(m.Key, m.Secret, ""),
			Secure: m.UseSSL,
		})
		if err != nil {
			log.Println(err)
		}
		m.client = client
	})
	return m.client
}

// Put transfers a file to the remote file system
func (m *Minio) Put(fileName, folder string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// build the object key without a leading slash: spec-correct S3 servers
	// (AWS, SeaweedFS, Garage) treat "/name" and "name" as different keys
	objectKey := strings.TrimPrefix(path.Join(folder, path.Base(fileName)), "/")
	client := m.getCredentials()
	uploadInfo, err := client.FPutObject(ctx, m.Bucket, objectKey, fileName, minio.PutObjectOptions{})
	if err != nil {
		log.Println("Failed with FPutObject")
		log.Println(err)
		log.Println("UploadInfo:", uploadInfo)
		return err
	}

	return nil
}

// List returns a listing of all files in the remote bucket with the
// given prefix, except for files with a leading . in the name
func (m *Minio) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	// object keys never start with "/"; a root listing on a spec-correct
	// S3 server needs an empty prefix, not "/"
	objectCh := client.ListObjects(ctx, m.Bucket, minio.ListObjectsOptions{
		Prefix:    strings.TrimPrefix(prefix, "/"),
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return listing, object.Err
		}

		if !strings.HasPrefix(object.Key, ".") {
			mb := filesystems.SizeToMB(object.Size)
			item := filesystems.Listing{
				Etag:         object.ETag,
				LastModified: object.LastModified,
				Key:          object.Key,
				Size:         mb,
			}
			listing = append(listing, item)
		}
	}

	return listing, nil
}

// Delete removes one or more files from the remote filesystem
func (m *Minio) Delete(itemsToDelete []string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

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

// Get pulls a file from the remote file system and saves it somewhere on our server
func (m *Minio) Get(destination string, items ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	for _, item := range items {
		err := client.FGetObject(ctx, m.Bucket, item, fmt.Sprintf("%s/%s", destination, path.Base(item)), minio.GetObjectOptions{})
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
