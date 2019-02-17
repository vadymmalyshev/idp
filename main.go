package main

import (
	"github.com/gin-gonic/gin"

	"git.tor.ph/hiveon/idp/auth"
	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models"
	ginutils "git.tor.ph/hiveon/idp/pkg/gin"
)

var (
	log = config.Logger()
)

func main() {
	r := gin.New()
	r.Use(ginutils.Middleware(log))

	db := config.DB()
	defer db.Close()

	models.Migrate(db)

	auth.Init(r)

	log.Infof("IDP has started on http://%s", config.ServerAddr)
	r.Run(config.ServerAddr)
}
