package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	serverPort = "idp.port"
	serverHost = "idp.host"

	cookieDomain = "cookie_domain"

	hydraAdmin        = "hydra.admin"
	hydraAPI          = "hydra.api"
	hydraClientID     = "hydra.client_id"
	hydraClientSecret = "hydra.client_secret"
	hydraIntrospect   = "hydra.introspect"

	dbHost     = "idp.db.host"
	dbPort     = "idp.db.port"
	dbName     = "idp.db.name"
	dbUser     = "idp.db.user"
	dbPassword = "idp.db.password"
	dbSSLMode  = "idp.db.sslmode"

	authSignKey = "auth.sign_key"

	mailFrom     = "mail.from"
	mailSMTP     = "mail.smtp"
	mailPort     = "mail.port"
	mailUser     = "mail.user"
	mailPassword = "mail.password"

	portalPort         = "portal.port"
	portalHost         = "portal.host"
	portalCallback     = "portal.callback"
	portalClientID     = "portal.client_id"
	portalClientSecret = "portal.client_secret"
)

var configName = flag.String("c", "config", "config file name from config directory")

func InitViperConfig() *CommonConfig {
	logrus.Info("THAT'S HAPPEN'")
	logrus.Infof("used config file: %s.yaml", *configName)
	// viper.AddConfigPath("$HOME/config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetConfigName(*configName)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	viper.SetDefault(serverHost, "localhost")
	viper.SetDefault(serverPort, "3000")

	viper.SetDefault(dbSSLMode, false)

	viper.SetDefault(authSignKey, "")

	viper.SetDefault(hydraAPI, "localhost:4444")
	viper.SetDefault(hydraAdmin, "localhost:4445")
	viper.SetDefault(hydraClientID, "")
	viper.SetDefault(hydraClientSecret, "")

	viper.SetDefault(cookieDomain, ".hiveon.localhost")

	var conf = new(CommonConfig)

	if err := viper.Unmarshal(conf, YAMLUnmarshalOpt); err != nil {
		log.Fatalln("Error while unmarshal viper config", err)
		return nil
	}

	conf.IDP.DB.Conn = getDBConn(conf.IDP)
	conf.ServerConfig, _ = initServerConfig(conf.IDP)

	return conf
}

type DBConfig struct {
	Conn     string
	Host     string
	Port     string
	SSLMode  string
	Name     string
	User     string
	Password string
}

// getDBConn returns db conn url
func getDBConn(conf IDP) string {
	sslMode := "disable"
	if conf.DB.Sslmode {
		sslMode = "enable"
	}
	return fmt.Sprintf("host=%s port=%s sslmode=%s user=%s dbname=%s password=%s ",
		conf.DB.Host, conf.DB.Port, sslMode, conf.DB.User, conf.DB.Name, conf.DB.Password)
}

func initServerConfig(conf IDP) (ServerConfig, error) {
	config := ServerConfig{
		Port: conf.Port,
		Host: conf.Host,
	}

	config.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)

	return config, nil
}

func GetMailConfig() (MailConfig, error) {
	config := MailConfig{
		From:     viper.GetString(mailFrom),
		SMTP:     viper.GetString(mailSMTP),
		Port:     viper.GetInt(mailPort),
		User:     viper.GetString(mailUser),
		Password: viper.GetString(mailPassword),
	}

	return config, nil
}

func GetCookieDomain() (string, error) {
	return viper.GetString(cookieDomain), nil
}

func YAMLUnmarshalOpt(c *mapstructure.DecoderConfig) {
	c.TagName = "yaml"
}
