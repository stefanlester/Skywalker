package webdavfilesystem

import "github.com/stefanlester/skywalker/filesystems"
// creating sub file systems for webdav
type WebDAV struct {
	Host string
	User string
	Pass string
}

func (s *WebDAV) Put(fileName, folder string) error {
	return nil
}

func (s *WebDAV) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	return listing, nil
}

func (s *WebDAV) Delete(itemsToDelete []string) bool {
	return true
}

func (s *WebDAV) Get(destination string, items ...string) error {
	return nil
}
