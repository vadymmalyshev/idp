package config

import (
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

var dbOnce sync.Once
var db *gorm.DB

func initDatabase() {
	var err error
	dbconfig, err := GetDBConfig()

	if err != nil {
		logrus.Panic("failed to find db config:", err.Error())
	}

	db, err = gorm.Open("postgres", dbconfig.Conn)

	db.LogMode(true)

	if err != nil {
		logrus.Panic("failed to init db:", err.Error())
	}

}

func DB() *gorm.DB {
	dbOnce.Do(initDatabase)

	return db
}
