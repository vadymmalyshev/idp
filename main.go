package main

import (
	"github.com/gin-contrib/static"
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

	db := config.DB()
	defer db.Close()

	models.Migrate(db)

	r.Use(ginutils.Middleware(log))
	r.Use(static.Serve("/assets", static.LocalFile("./views/assets", true)))

	auth.Init(r, db)

	serverConfig, _ := config.GetServerConfig()

	log.Infof("IDP has started on http://%s", serverConfig.Addr)

	r.Run(serverConfig.Addr)
}
