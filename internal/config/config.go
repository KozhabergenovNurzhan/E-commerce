package config

type Config struct {
	Port     string
	LogLevel string
	DBConfig *DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}
