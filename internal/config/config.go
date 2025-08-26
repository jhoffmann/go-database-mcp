package config

type Config struct {
	Database DatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
	Type         string `json:"type" envconfig:"DB_TYPE"`
	Host         string `json:"host" envconfig:"DB_HOST"`
	Port         int    `json:"port" envconfig:"DB_PORT"`
	Database     string `json:"database" envconfig:"DB_NAME"`
	Username     string `json:"username" envconfig:"DB_USER"`
	Password     string `json:"password" envconfig:"DB_PASSWORD"`
	MaxConns     int    `json:"max_conns" envconfig:"DB_MAX_CONNS"`
	MaxIdleConns int    `json:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS"`
	SSLMode      string `json:"ssl_mode" envconfig:"DB_SSL_MODE"`
}
