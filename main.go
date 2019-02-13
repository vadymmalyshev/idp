package main

import (
	"idp/config"

	"git.tor.ph/idp/hiveon/internal/gin"

	"github.com/gin-gonic/gin"
)

var (
	logger = config.Logger()
)

func main() {
	db := config.DB()

	r := gin.New()
	r.Use(ginutils.Middleware(logger))
}
