package config

import (
	"sync"

	"github.com/jinzhu/gorm"
)

var dbOnce sync.Once
var db *gorm.DB

func initDatabase() {
	var err error
	db, err = gorm.Open("postgres", DBConn)

	db.LogMode(true)

	logger := Logger()

	if err != nil {
		logger.Panic("failed to init db:", err.Error())
	}
}

func DB() *gorm.DB {
	dbOnce.Do(initDatabase)

	return db
}
