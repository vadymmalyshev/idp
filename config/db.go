package config

import (
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

var dbOnce sync.Once
var db *gorm.DB

func initDatabase(dbConf DBConf) {
	db, err := gorm.Open("postgres", dbConf.Conn)

	db.LogMode(true)

	if err != nil {
		logrus.Panic("failed to init db:", err.Error())
	}

}

func DB(dbConf DBConf) *gorm.DB {
	dbOnce.Do(func() {
		initDatabase(dbConf)
	})

	return db
}
