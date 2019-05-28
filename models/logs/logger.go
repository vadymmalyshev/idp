package logs

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var (
	assertLog   = &Log{}
	assertLogger = &UserLogger{}
)

type UserLogger struct {
	db *gorm.DB
}

func NewUserLogger(db *gorm.DB) *UserLogger {
	return &UserLogger{db}
}

func (logger UserLogger) CreateRecord(log *Log) error {
	err := logger.db.Create(log).Error
	if err != nil {
		return errors.New("Can't create log record")
	}

	logrus.Infof("Record created for user: ", logrus.Fields{
		"UserID": log.UserID,
	})

	return err
}

func (logger UserLogger) New() *Log {
	return &Log{}
}