package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	serverPort = "idp.port"
	serverHost = "idp.host"

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

func init() {
	logrus.Info("THAT'S HAPPEN'")
	// viper.AddConfigPath("$HOME/config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetConfigName("config")

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

// GetDBConfig returns db config
// todo error checking
func GetDBConfig() (DBConfig, error) {
	config := DBConfig{
		Host:     viper.GetString(dbHost),
		Port:     viper.GetString(dbPort),
		Name:     viper.GetString(dbName),
		User:     viper.GetString(dbUser),
		Password: viper.GetString(dbPassword),
	}

	if viper.GetBool(dbSSLMode) {
		config.SSLMode = "enable"
	} else {
		config.SSLMode = "disable"
	}

	config.Conn = fmt.Sprintf("host=%s port=%s sslmode=%s user=%s dbname=%s password=%s ",
		config.Host, config.Port, config.SSLMode, config.User, config.Name, config.Password)

	return config, nil
}

type ServerConfig struct {
	Addr string
	Host string
	Port string
}

func GetServerConfig() (ServerConfig, error) {
	config := ServerConfig{
		Port: viper.GetString(serverPort),
		Host: viper.GetString(serverHost),
	}

	config.Addr = fmt.Sprintf("%s:%s", config.Host, config.Port)

	return config, nil
}

func GetSignKey() (string, error) {
	key := viper.GetString(authSignKey)

	if key == "" {
		return key, errors.New("Token signing key is missing from configuration")
	}
	if len(key) < 32 {
		return key, errors.New("Token signing key must be at least 32 characters")
	}

	return key, nil
}

type HydraConfig struct {
	Admin        string
	API          string
	ClientID     string
	ClientSecret string
	Introspect   string
}

func GetHydraConfig() (*HydraConfig, error) {
	config := HydraConfig{
		Admin:        viper.GetString(hydraAdmin),
		API:          viper.GetString(hydraAPI),
		ClientID:     viper.GetString(hydraClientID),
		ClientSecret: viper.GetString(hydraClientSecret),
		Introspect:   viper.GetString(hydraIntrospect),
	}

	return &config, nil
}

type MailConfig struct {
	From     string
	SMTP     string
	Port     int
	User     string
	Password string
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

type PortalConfig struct {
	Port         int
	Host         string
	Callback     string
	ClientID     string
	ClientSecret string
}

func GetPortalConfig() (PortalConfig, error) {
	config := PortalConfig{
		Port:         viper.GetInt(portalPort),
		Host:         viper.GetString(portalHost),
		Callback:     viper.GetString(portalCallback),
		ClientID:     viper.GetString(portalClientID),
		ClientSecret: viper.GetString(portalClientSecret),
	}

	return config, nil
}
