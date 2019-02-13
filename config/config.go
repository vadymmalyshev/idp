package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	idpPort = "idp.port"
	idpHost = "idp.host"

	hydraAdmin        = "hydra.admin"
	hydraAPI          = "hydra.api"
	hydraClientID     = "hydra.client_id"
	hydraClientSecret = "hydra.client_secret"

	dbHost     = "db.host"
	dbPort     = "db.port"
	dbName     = "db.name"
	dbUser     = "db.user"
	dbPassword = "db.password"
	dbSSLMode  = "db.sslmode"
)

var (
	ServerPort, ServerHost string

	DBConn string

	HydraAdmin, HydraAPI, HydraClientID, HydraClientSecret string
)

func init() {
	// viper.AddConfigPath("$HOME/config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetConfigName("config")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	ServerHost = viper.GetString("server.host")
	ServerPort = viper.GetString("server.port")

	sslmode := "disable"
	isSsl := viper.GetBool("db.sslmode")
	if isSsl {
		sslmode = "enable"
	}

	DBConn = fmt.Sprintf("host=%s port=%s sslmode=%s user=%s dbname=%s password=%s ",
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		sslmode,
		viper.GetString("db.user"),
		viper.GetString("db.name"),
		viper.GetString("db.password"),
	)

	HydraAdmin = viper.GetString("hydra.admin")
	HydraAPI = viper.GetString("hydra.api")
	HydraClientID = viper.GetString("hydra.client_id")
	HydraClientSecret = viper.GetString("hydra.client_secret")

}
