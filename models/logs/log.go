package logs

import (
	"github.com/jinzhu/gorm"
	"time"
)
// login/registration log, type - login/registration

type Log struct {
	gorm.Model
	UserID string  `gorm:"not null"`
	Agent  string  `gorm:"not null"`
	IP     string  `gorm:"not null"`
	Domen  string  `gorm:"not null"`
	Type   string  `gorm:"not null"`
}

// TableName represents gorm interface to change users table name
func (Log) TableName() string {
	return "user_log"
}

func (l *Log) PutUserID(userID string) { l.UserID = userID }

func (l *Log) PutAgent(agent string) { l.Agent = agent }

func (l *Log) PutIP(ip string) { l.IP = ip }

func (l *Log) PutDomen(domen string) { l.Domen = domen }

func (l *Log) GetUserID() string { return l.UserID}

func (l *Log) GetAgent() string { return l.Agent}

func (l *Log) GetIP() string { return l.IP}

func (l *Log) GetDomen() string { return l.Domen}

func (l *Log) GetDate() time.Time { return l.CreatedAt}