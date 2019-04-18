package auth

import (
	"git.tor.ph/hiveon/idp/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Auth struct {
	r    *gin.Engine
	db   *gorm.DB
	conf *config.CommonConfig
}

func NewAuth(r *gin.Engine, db *gorm.DB, conf *config.CommonConfig) *Auth {
	return &Auth{
		r:    r,
		db:   db,
		conf: conf,
	}
}


func (a *Auth) Init() {

}
