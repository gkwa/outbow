package outbow

import (
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	mutex sync.Mutex
)

// DatabaseStorage is an implementation of URLStorage using a database.
type DatabaseStorage struct{}

func (d *DatabaseStorage) SaveURL(url string) error {
	mutex.Lock()
	defer mutex.Unlock()

	// Implementation for saving to the database (GORM, in this case)
	newURL := URL{URL: url, CreatedAt: time.Now()}
	result := db.Create(&newURL)

	return result.Error
}

func (d *DatabaseStorage) LoadURLs() (map[string]time.Time, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var urls []URL
	result := db.Find(&urls)
	if result.Error != nil {
		return nil, result.Error
	}

	urlsMap := make(map[string]time.Time)
	for _, u := range urls {
		urlsMap[u.URL] = u.CreatedAt
	}

	return urlsMap, nil
}

func (d *DatabaseStorage) IsURLPresent(url string) (bool, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var count int64
	result := db.Model(&URL{}).Where("url = ?", url).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func initializeDB() error {
	var err error
	db, err = gorm.Open(sqlite.Open("urls.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	// AutoMigrate creates the URL table if it doesn't exist
	err = db.AutoMigrate(&URL{})
	return err
}
