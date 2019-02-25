package main

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gogap/logrus_mate"

	"git.tor.ph/hiveon/idp/auth"
	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/models"
	ginutils "git.tor.ph/hiveon/idp/pkg/gin"
	"git.tor.ph/hiveon/idp/pkg/log"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus_mate.Hijack(
		logrus.StandardLogger(),
		logrus_mate.ConfigString(
			`{formatter.name = "text"}`,
		),
	)
}

func main() {

	r := gin.New()

	db := config.DB()
	defer db.Close()

	models.Migrate(db)

	logger := log.NewLogger(log.Config{
		Level:  "debug",
		Format: "text",
	})

	r.Use(ginutils.Middleware(logger))

	r.Use(static.Serve("/assets", static.LocalFile("./views/assets", true)))

	auth.Init(r, db)

	serverConfig, _ := config.GetServerConfig()

	logrus.Infof("IDP has started on http://%s", serverConfig.Addr)

	r.Run(serverConfig.Addr)
}
