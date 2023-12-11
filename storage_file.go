package outbow

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"
)

type FileStorage struct {
	FileName string
}

func (f *FileStorage) SaveURL(url string) error {
	urls, err := f.loadFromFile()
	if err != nil {
		return err
	}

	if _, exists := urls[url]; exists {
		slog.Debug("URL already exists, skipping", "url", url)
		return nil
	}

	urls[url] = time.Now()

	return f.saveToFile(urls)
}

func (f *FileStorage) LoadURLs() (map[string]time.Time, error) {
	return f.loadFromFile()
}

func (f *FileStorage) IsURLPresent(url string) (bool, error) {
	urls, err := f.loadFromFile()
	if err != nil {
		return false, err
	}

	_, exists := urls[url]
	return exists, nil
}

func (f *FileStorage) loadFromFile() (map[string]time.Time, error) {
	file, err := os.Open(f.FileName)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]time.Time), nil
		}
		return nil, err
	}
	defer file.Close()

	var urls map[string]time.Time
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&urls)
	if err != nil {
		if err == io.EOF {
			return make(map[string]time.Time), nil
		}
		return nil, err
	}

	return urls, nil
}

func (f *FileStorage) saveToFile(urls map[string]time.Time) error {
	file, err := os.Create(f.FileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encodedData, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(encodedData)
	if err != nil {
		return err
	}

	return nil
}
