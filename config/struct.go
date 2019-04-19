package config

type HydraConfig struct {
	Admin        string `yaml:"admin"`
	API          string `yaml:"api"`
	Docker       string `yaml:"docker"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Introspect   string `yaml:"introspect"`
}

type IDP struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Callback     string `yaml:"callback"`
	DB           DBConf `yaml:"db"`
}

type DBConf struct {
	Conn    string `yaml:"-"`
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	Name    string `yaml:"name"`
	User    string `yaml:"user"`
	Sslmode bool   `yaml:"sslmode"`

	Password string `yaml:"password"`
}

type PortalConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	Callback     string `yaml:"callback"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type AdminConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Callback     string `yaml:"callback"`
	DB           DBConf `yaml:"db"`
}

type MailConfig struct {
	Active   string `yaml:"active"`
	SMTP     string `yaml:"smtp"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Db       int    `yaml:"db"`
	Password string `yaml:"password"`
}

type CommonConfig struct {
	Hydra  HydraConfig  `yaml:"hydra"`
	IDP    IDP          `yaml:"idp"`
	Portal PortalConfig `yaml:"portal"`
	Admin  AdminConfig  `yaml:"admin"`
	Auth   struct {
		SignKey string `yaml:"sign_key"`
	} `yaml:"auth"`
	Mail         MailConfig   `yaml:"mail"`
	Redis        RedisConfig  `yaml:"redis"`
	RememberFor  int          `yaml:"remember_for"`
	CookieDomain string       `yaml:"cookie_domain"`
	ServerConfig ServerConfig `yaml:"-"`
}

type ServerConfig struct {
	Addr string
	Host string
	Port int
}
