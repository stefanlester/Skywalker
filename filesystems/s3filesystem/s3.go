package s3filesystem

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stefanlester/skywalker/filesystems"
)

// S3 is the overall type for the s3 filesystem, and contains the connection
// credentials, region, endpoint, and the bucket to use. It is backed by the
// minio-go client, which is S3-compatible and works against AWS S3.
type S3 struct {
	Key      string
	Secret   string
	Region   string
	Endpoint string
	Bucket   string
}

// getCredentials generates a minio client using the credentials stored in
// the S3 type. S3 always uses HTTPS, so Secure is hard-coded to true.
func (s *S3) getCredentials() *minio.Client {
	client, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.Key, s.Secret, ""),
		Region: s.Region,
		Secure: true,
	})
	if err != nil {
		log.Println(err)
	}
	return client
}

// Put transfers a file to the remote file system
func (s *S3) Put(fileName, folder string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectName := path.Base(fileName)
	client := s.getCredentials()
	uploadInfo, err := client.FPutObject(ctx, s.Bucket, fmt.Sprintf("%s/%s", folder, objectName), fileName, minio.PutObjectOptions{})
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
func (s *S3) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := s.getCredentials()

	objectCh := client.ListObjects(ctx, s.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return listing, object.Err
		}

		if !strings.HasPrefix(object.Key, ".") {
			b := float64(object.Size)
			kb := b / 1024
			mb := kb / 1024
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
func (s *S3) Delete(itemsToDelete []string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := s.getCredentials()

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}

	for _, item := range itemsToDelete {
		err := client.RemoveObject(ctx, s.Bucket, item, opts)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	return true
}

// Get pulls a file from the remote file system and saves it somewhere on our server
func (s *S3) Get(destination string, items ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := s.getCredentials()

	for _, item := range items {
		err := client.FGetObject(ctx, s.Bucket, item, fmt.Sprintf("%s/%s", destination, path.Base(item)), minio.GetObjectOptions{})
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
