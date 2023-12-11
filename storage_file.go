package outbow

import (
	"time"
)

// FileStorage is an implementation of URLStorage using a file.
type FileStorage struct {
	FileName string
}

func (f *FileStorage) SaveURL(url string) error {
	return nil
}

func (f *FileStorage) LoadURLs() (map[string]time.Time, error) {
	// Implementation for loading from a file
	// ...

	return nil, nil
}

func (f *FileStorage) IsURLPresent(url string) (bool, error) {
	// Implementation for checking URL existence in the file
	// ...

	return false, nil
}
