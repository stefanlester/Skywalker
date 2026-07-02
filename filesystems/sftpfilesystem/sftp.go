package sftpfilesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/stefanlester/skywalker/filesystems"
	"golang.org/x/crypto/ssh"
)

// SFTP is the overall type for the sftp filesystem, and contains the
// connection credentials (host, user, password) and the port to connect on.
type SFTP struct {
	Host string
	User string
	Pass string
	Port string
}

// getConnection dials the remote host over SSH using password authentication
// and returns an sftp client along with the underlying ssh connection, so the
// caller can close both. HostKeyCallback is set to InsecureIgnoreHostKey, which
// is acceptable for this demo backend but should be replaced with a known_hosts
// callback in production.
func (s *SFTP) getConnection() (*sftp.Client, *ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", s.Host, s.Port), config)
	if err != nil {
		return nil, nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	return client, conn, nil
}

// Put transfers a file to the remote file system, storing it under folder using
// the base name of fileName
func (s *SFTP) Put(fileName, folder string) error {
	client, conn, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return err
	}
	defer client.Close()
	defer conn.Close()

	src, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return err
	}
	defer src.Close()

	dst, err := client.Create(fmt.Sprintf("%s/%s", folder, path.Base(fileName)))
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
}

// List returns a listing of all files in the remote directory named by prefix,
// except for files with a leading . in the name
func (s *SFTP) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	client, conn, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return listing, err
	}
	defer client.Close()
	defer conn.Close()

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
func (s *SFTP) Delete(itemsToDelete []string) bool {
	client, conn, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return false
	}
	defer client.Close()
	defer conn.Close()

	for _, item := range itemsToDelete {
		if err := client.Remove(item); err != nil {
			log.Println(err)
			return false
		}
	}

	return true
}

// Get pulls one or more files from the remote file system and saves them under
// destination on our server, using the base name of each item
func (s *SFTP) Get(destination string, items ...string) error {
	client, conn, err := s.getConnection()
	if err != nil {
		log.Println(err)
		return err
	}
	defer client.Close()
	defer conn.Close()

	for _, item := range items {
		err := func() error {
			src, err := client.Open(item)
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
