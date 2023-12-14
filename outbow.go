package outbow

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
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

	slog.Debug("stats",
		"pageCount", maxPageNumber,
		"reviewCount", reviewCount,
		"reviewsPerPage", reviewsPerPage,
		"quotient", quotient,
		"remainder", remainder,
	)

	return maxPageNumber
}

type URLStorage interface {
	SaveURL(url string) error
	LoadURLs() (map[string]time.Time, error)
	IsURLPresent(url string) (bool, error)
}

const (
	DataDir               = "data"
	numberFormatSpecifier = "%04d"
)

var (
	storage        URLStorage
	DataDirAbsPath string
)

func init() {
	var err error
	DataDirAbsPath, err = filepath.Abs(DataDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

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
			Path:   "/", // Default value
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

	sites := []GoProModelSite{
		NewGoProModelSite("max", WithReviewCount(325), WithPageBasePath("/en/us/shop/cameras/max/CHDHZ-202-master.html")),
		NewGoProModelSite("Hero10", WithReviewCount(2373), WithPageBasePath("/en/us/shop/cameras/hero10-black/CHDHX-101-master.html")),
		NewGoProModelSite("Hero11", WithReviewCount(1358), WithPageBasePath("/en/us/shop/cameras/hero11-black/CHDHX-111-master.html")),
		NewGoProModelSite("Hero12", WithReviewCount(118), WithPageBasePath("/en/us/shop/cameras/hero12-black/CHDHX-121-master.html")),
	}

	for _, site := range sites {
		err = dowork(site, storedURLs, options)
		if err != nil {
			slog.Error("dowork", "err", err)
			return 1
		}
	}

	return 0
}

func dowork(site GoProModelSite, storedURLs map[string]time.Time, options options.Options) error {
	urlCreationStrategy := DefaultURLCreationStrategy{}

	maxPageNumber := site.TotalPageCount()
	pageNumbers := barpear.RandomPositiveIntegerSliceUpToMax(maxPageNumber)
	baseURL := site.HomePage // first reviews start at product homepage

	slog.Debug("page numbers",
		"count", len(pageNumbers),
		"slice", pageNumbers,
	)

	var allPages []PageNumberContainer
	for pageNum := range pageNumbers {
		if pageNum == 0 { // FIXME
			continue
		}
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
	slog.Debug("subset", "url", slog.Int("subset", options.SubsetPercentage))
	if options.SubsetPercentage < 100 {
		slog.Warn("subset", "url limited", slog.Int("subset", options.SubsetPercentage))
	}
	maxIndex := len(pagesNotYetFetched) * options.SubsetPercentage / 100

	slog.Debug("subtask limit",
		"maxIndex", maxIndex,
		"allPages count", len(allPages),
		"remaining", len(pagesNotYetFetched),
	)

	willFetch := pagesNotYetFetched[:maxIndex]

	slog.Debug("willfetch order",
		"count", len(willFetch),
		"willfetch", willFetch,
	)

	for pageIter := 0; pageIter < len(willFetch); pageIter++ {
		page := willFetch[pageIter]
		url := page.URL

		slog.Debug("fetching this run",
			"remaining", len(pagesNotYetFetched)-pageIter,
			"this batch size", len(willFetch),
			"remaining this run", len(willFetch)-pageIter,
		)

		osascript := OsaScript{
			PageNumberContainer:     page,
			AllowReviewsLoadSeconds: options.AllowReviewsLoadSeconds,
		}

		// don't re-fetch
		_, found := storedURLs[page.URL.String()]
		if found {
			slog.Debug("skipping refetch", "url", url.String())
			continue
		}

		slog.Debug("page", "number", page.PageNumber)

		osascript.WriteApplescript(site.Model)

		if options.NoRunOsascript {
			continue
		}

		slog.Debug("command", "command to run", osascript.CommandResult.CommandString())

		// clear clipboard
		clipboard.WriteAll("")

		// fill clipboard
		err := osascript.CommandResult.Run()
		if err != nil {
			slog.Error("command run", "error", err)
			return err
		}

		// read clipboard into var
		clipboardContent, err := clipboard.ReadAll()
		if err != nil {
			slog.Error("reading clipboard", "error", err)
			return err
		}

		// write clipboard to data file
		y := "gopro-%s-" + numberFormatSpecifier + ".txt"
		outFname := fmt.Sprintf(y, strings.ToLower(site.Model), page.PageNumber)
		outPath := filepath.Join(DataDirAbsPath, outFname)

		slog.Debug("writing data", "path", outPath)

		err = os.MkdirAll(DataDirAbsPath, os.ModePerm)
		if err != nil {
			slog.Error("mkdir had error", "dir", DataDirAbsPath, "error", err)
			return err
		}

		if err := os.WriteFile(outPath, []byte(clipboardContent), 0o600); err != nil {
			fmt.Println("Error:", err)
			return err
		}

		if err := storage.SaveURL(url.String()); err != nil {
			slog.Error("error saving urls", "error", err)
			return err
		}
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
