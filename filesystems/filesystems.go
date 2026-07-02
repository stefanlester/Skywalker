package filesystems

import "time"

// FS is the interface for file systems. In order to satisfy the interface,
// all of its functions must exist
type FS interface {
	Put(fileName, folder string) error
	Get(destination string, items ...string) error
	List(prefix string) ([]Listing, error)
	Delete(itemsToDelete []string) bool
}

// Listing describes one file on a remote file system
type Listing struct {
	Etag         string
	LastModified time.Time
	Key          string
	Size         float64
	IsDir        bool
}

// SizeToMB converts a size in bytes to megabytes, as reported in Listing.Size
func SizeToMB(size int64) float64 {
	return float64(size) / 1024 / 1024
}
