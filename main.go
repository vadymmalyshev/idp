package main

import (
	"github.com/gin-gonic/gin"

	"git.tor.ph/hiveon/idp/auth"
	"git.tor.ph/hiveon/idp/config"
	ginutils "git.tor.ph/hiveon/idp/internal/gin"
)

var (
	log = config.Logger()
)

func main() {
	r := gin.New()
	r.Use(ginutils.Middleware(log))

	db := config.DB()
	defer db.Close()

	auth.Init(r)

	log.Infof("IDP has started on http://%s", config.ServerAddr)
	r.Run(config.ServerAddr)
}
