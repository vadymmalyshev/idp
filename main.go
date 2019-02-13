package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"git.tor.ph/hiveon/idp/auth"
	"git.tor.ph/hiveon/idp/config"
	"git.tor.ph/hiveon/idp/internal/gin"
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

	errs := make(chan error, 2)

	go func() {
		log.WithFields(
			logrus.Fields{
				"transport:": "http",
				"port":       config.ServerPort,
			}).Info("IDP has started")

		errs <- r.Run(":" + config.ServerPort)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	log.Info("terminated", <-errs)
}
