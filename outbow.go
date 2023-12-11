package outbow

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"net/url"
	"os"
	"text/template"
	"time"
)

func Main(storageType string) int {
	var storage URLStorage

	switch storageType {
	case "db":
		storage = &DatabaseStorage{FileName: "urls.db"}
		InitializeStorage(storage)
	case "json":
		storage = &FileStorage{FileName: "urls.json"}
		InitializeStorage(storage)
	default:
		slog.Error("invalid storage type. Supported values: db, json")
		return 1
	}
	slog.Debug("outbow", "test", true)

	InitializeStorage(storage)

	storedURLs, err := storage.LoadURLs()
	if err != nil {
		slog.Error("loading urls", "error", err)
	}

	for key, value := range storedURLs {
		slog.Debug("debug", "url", key, "fetch time", value.Format(time.RFC3339))
	}

	reviewCount := 1358
	reviewsPerPage := 5
	quotient := reviewCount / reviewsPerPage
	remainder := reviewCount % reviewsPerPage
	var maxPageNumber int64 = int64(quotient + 1)

	slog.Debug("stats", "pageCount", maxPageNumber, "reviewCount", reviewCount, "reviewsPerPage", reviewsPerPage, "quotient", quotient, "remainder", remainder)
	n, err := rand.Int(rand.Reader, big.NewInt(maxPageNumber+1))
	if err != nil {
		slog.Error("error generating random number", err)
	}

	pageCount := int(n.Int64())
	myURL := createURL(pageCount)

	if err := storage.SaveURL(myURL.String()); err != nil {
		slog.Error("error saving urls", "error", err)
	}

	slog.Debug("stats", "pageCount", maxPageNumber, "reviewCount", reviewCount, "reviewsPerPage", reviewsPerPage, "quotient", quotient, "remainder", remainder)

	// don't re-fetch
	_, found := storedURLs[myURL.String()]
	if found {
		slog.Debug("skipping refetch", "url", myURL.String())
		slog.Debug("stats", "pageCount", maxPageNumber, "reviewCount", reviewCount, "reviewsPerPage", reviewsPerPage, "quotient", quotient, "remainder", remainder)
		return 0
	}

	var applescriptBuf bytes.Buffer
	if err := genApplescript(&applescriptBuf, pageCount, myURL); err != nil {
		slog.Error("generating Applescript", "error", err)
		return 1
	}

	fname := fmt.Sprintf("gopro%04d.scpt", pageCount)

	if err := writeToFile(fname, applescriptBuf.Bytes()); err != nil {
		slog.Error("writing applescript to file", "error", err)
		return 1
	}

	return 0
}

type URLStorage interface {
	SaveURL(url string) error
	LoadURLs() (map[string]time.Time, error)
	IsURLPresent(url string) (bool, error)
}

var storage URLStorage

type URL struct {
	URL       string
	CreatedAt time.Time
}

func genApplescript(outputBuffer *bytes.Buffer, pageCount int, myURL *url.URL) error {
	tmpl, err := template.ParseFiles("gopro.scpt.tmpl")
	if err != nil {
		return fmt.Errorf("error reading template: %v", err)
	}

	data := struct {
		MyURL string
	}{
		MyURL: myURL.String(),
	}

	if err := tmpl.Execute(outputBuffer, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}

func createURL(pageCount int) *url.URL {
	path := "/en/us/shop/cameras/hero11-black/CHDHX-111-master.html"
	baseURL := &url.URL{
		Scheme: "https",
		Host:   "gopro.com",
		Path:   path,
	}

	if pageCount == 1 {
		return baseURL
	}

	baseURL.RawQuery = fmt.Sprintf("yoReviewsPage=%d", pageCount)
	return baseURL
}

func writeToFile(filename string, content []byte) error {
	outputFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.Write(content); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func InitializeStorage(s URLStorage) {
	storage = s

	if _, ok := storage.(*DatabaseStorage); ok {
		dbFname := storage.(*DatabaseStorage).FileName
		if err := initializeDB(dbFname); err != nil {
			slog.Error("error initializing database", err)
		}
	}
}
