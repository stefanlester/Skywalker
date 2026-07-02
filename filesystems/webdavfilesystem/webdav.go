package webdavfilesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/stefanlester/skywalker/filesystems"
	"github.com/studio-b12/gowebdav"
)

// WebDAV is the overall type for the webdav filesystem, and contains the
// connection credentials (host, user, password) used to reach the server.
type WebDAV struct {
	Host string
	User string
	Pass string

	clientMu sync.Mutex
	client   *gowebdav.Client
}

// getConnection returns a gowebdav client for the stored credentials. The
// client is built and its connectivity verified via Connect once, then cached
// for reuse (gowebdav clients are safe for concurrent use). Connect performs
// network I/O, so a failure is not cached: the next call retries.
func (s *WebDAV) getConnection() (*gowebdav.Client, error) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	client := gowebdav.NewClient(s.Host, s.User, s.Pass)
	if err := client.Connect(); err != nil {
		return nil, err
	}

	s.client = client
	return s.client, nil
}

// Put transfers a file to the remote file system, storing it under folder using
// the base name of fileName. The local file is streamed to the server rather
// than read fully into memory.
func (s *WebDAV) Put(fileName, folder string) error {
	client, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return err
	}

	src, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return err
	}
	defer src.Close()

	if err := client.WriteStream(path.Join(folder, path.Base(fileName)), src, 0644); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// List returns a listing of all files in the remote directory named by prefix,
// except for files with a leading . in the name
func (s *WebDAV) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	client, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return listing, err
	}

	files, err := client.ReadDir(prefix)
	if err != nil {
		log.Println(err)
		return listing, err
	}

	for _, x := range files {
		if !strings.HasPrefix(x.Name(), ".") {
			mb := filesystems.SizeToMB(x.Size())
			item := filesystems.Listing{
				Key:          x.Name(),
				Size:         mb,
				LastModified: x.ModTime(),
				IsDir:        x.IsDir(),
			}
			listing = append(listing, item)
		}
	}

	return listing, nil
}

// Delete removes one or more files from the remote filesystem, returning false
// as soon as any removal fails
func (s *WebDAV) Delete(itemsToDelete []string) bool {
	client, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return false
	}

	for _, item := range itemsToDelete {
		if err := client.Remove(item); err != nil {
			log.Println(err)
			return false
		}
	}

	return true
}

// Get pulls one or more files from the remote file system and saves them under
// destination on our server, using the base name of each item. Each item is
// streamed to disk rather than read fully into memory.
func (s *WebDAV) Get(destination string, items ...string) error {
	client, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return err
	}

	for _, item := range items {
		err := func() error {
			src, err := client.ReadStream(item)
			if err != nil {
				log.Println(err)
				return err
			}
			defer src.Close()

			dst, err := os.Create(fmt.Sprintf("%s/%s", destination, path.Base(item)))
			if err != nil {
				log.Println(err)
				return err
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				log.Println(err)
				return err
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
