package skywalker

import (
	"crypto/rand"
	"os"
)

const (
	randomString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// gnerates a random string length n from values in the const randomString
func (c *Skywalker) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomString)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}

// CreateDirIf NotExsit creates a new directory if it doesn't exist.
func (c *Skywalker) CreateDirIfNotExist(path string) error {
	const mode = 0755

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFileIfNotExist creates a new file at path if it does not exist.
func (c *Skywalker) CreateFileIfNotExist(path string) error {
	var _, err = os.Stat(path)

	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}

	return nil
}
