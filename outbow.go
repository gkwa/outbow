package outbow

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/taylormonacelli/barpear"
	"github.com/taylormonacelli/outbow/options"
)

type GoProModelSite struct {
	HomePage       url.URL
	Model          string
	ReviewCount    int
	ReviewsPerPage int
}

type PageNumberContainer struct {
	URL        *url.URL
	PageNumber int
}

func (site *GoProModelSite) TotalPageCount() int {
	reviewCount := site.ReviewCount
	reviewsPerPage := site.ReviewsPerPage

	quotient := reviewCount / reviewsPerPage
	remainder := reviewCount % reviewsPerPage

	maxPageNumber := quotient + 1
	if remainder == 0 {
		maxPageNumber = quotient
	}

	slog.Debug("stats", "pageCount", maxPageNumber, "reviewCount", reviewCount, "reviewsPerPage", reviewsPerPage, "quotient", quotient, "remainder", remainder)

	return maxPageNumber
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

type URLCreationStrategy interface {
	CreateURL(site GoProModelSite) *url.URL
}

type DefaultURLCreationStrategy struct{}

func NewGoProModelSite(model string, options ...func(*GoProModelSite)) GoProModelSite {
	site := GoProModelSite{
		Model:          model,
		ReviewCount:    0, // Default value
		ReviewsPerPage: 5, // Default value
		HomePage: url.URL{
			Scheme: "https",
			Host:   "gopro.com",
			Path:   "/",
		},
	}

	for _, option := range options {
		option(&site)
	}

	return site
}

func WithReviewCount(count int) func(*GoProModelSite) {
	return func(s *GoProModelSite) {
		s.ReviewCount = count
	}
}

func WithPageBasePath(path string) func(*GoProModelSite) {
	return func(s *GoProModelSite) {
		s.HomePage.Path = path
	}
}

func (s DefaultURLCreationStrategy) GenerateURL(baseURL url.URL, pageNum int) url.URL {
	if pageNum <= 1 {
		return baseURL
	}

	baseURL.RawQuery = fmt.Sprintf("yoReviewsPage=%d", pageNum)
	return baseURL
}

func Main(options options.Options) int {
	var storage URLStorage
	switch options.StorageType {
	case "db":
		storage = &DatabaseStorage{FileName: "urls.db"}
	case "json":
		storage = &FileStorage{FileName: "urls.json"}
	default:
		slog.Error("invalid storage type. Supported values: db, json")
		return 1
	}

	InitializeStorage(storage)

	storedURLs, err := storage.LoadURLs()
	if err != nil {
		slog.Error("loading urls", "error", err)
		return 1
	}

	for key, value := range storedURLs {
		slog.Debug("debug", "url", key, "fetch time", value.Format(time.RFC3339))
	}

	urlCreationStrategy := DefaultURLCreationStrategy{}

	reviewCount := 1358 // get this manually by visiting site to find review count
	site := NewGoProModelSite(
		"Hero11",
		WithReviewCount(reviewCount),
		WithPageBasePath("/en/us/shop/cameras/hero11-black/CHDHX-111-master.html"),
	)

	maxPageNumber := site.TotalPageCount()
	pageNumbers := barpear.RandomPositiveIntegerSliceUpToMax(maxPageNumber)
	baseURL := site.HomePage // first reviews start at product homepage

	var allPages []PageNumberContainer
	for pageNum := range pageNumbers {
		url := urlCreationStrategy.GenerateURL(baseURL, pageNum)
		pc := PageNumberContainer{URL: &url, PageNumber: pageNum}
		allPages = append(allPages, pc)
	}

	var pagesNotYetFetched []PageNumberContainer
	for _, page := range allPages {
		present, err := storage.IsURLPresent(page.URL.String())
		if err != nil {
			panic(err)
		}

		if !present {
			pagesNotYetFetched = append(pagesNotYetFetched, page)
		}
	}

	// create small subset in order to prevent overloading site
	y := len(allPages) * options.SubsetPercentage / 100
	z := len(pagesNotYetFetched)
	if z > 0 {
		z--
	}
	maxIndex := min(y, z)

	slog.Debug("subtask limit",
		"maxIndex", maxIndex,
		"y", y,
		"z", z,
		"remaining", len(pagesNotYetFetched),
	)

	willFetch := pagesNotYetFetched[:maxIndex]

	for _, page := range willFetch {
		url := page.URL
		// don't re-fetch
		_, found := storedURLs[page.URL.String()]
		if found {
			slog.Debug("skipping refetch", "url", url.String())
			continue
		}
		if err := storage.SaveURL(url.String()); err != nil {
			slog.Error("error saving urls", "error", err)
		}

		var applescriptBuf bytes.Buffer
		if err := genApplescript(&applescriptBuf, page.PageNumber, *page.URL); err != nil {
			slog.Error("generating Applescript", "error", err)
			return 1
		}

		fname := fmt.Sprintf("gopro%04d.scpt", page.PageNumber)
		if err := writeToFile(fname, applescriptBuf.Bytes()); err != nil {
			slog.Error("writing applescript to file", "error", err)
			return 1
		}

		time.Sleep(10 * time.Millisecond)
	}
	return 0
}

func genApplescript(outputBuffer *bytes.Buffer, pageCount int, myURL url.URL) error {
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
