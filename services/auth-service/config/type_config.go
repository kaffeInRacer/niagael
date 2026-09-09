package config

import "time"

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Logger   LoggerConfig   `yaml:"log"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type KafkaConfig struct {
	Brokers     []string `yaml:"brokers"`
	CasbinTopic string   `yaml:"casbin_topic"`
	UserTopic   string   `yaml:"user_topic"`
}

type HTTPConfig struct {
	ServiceName     string        `yaml:"service_name"`
	Mode            string        `yaml:"mode"`
	Addr            string        `yaml:"addr"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Stdout bool   `yaml:"stdout"`
	Path   string `yaml:"path"`
}

type PostgresConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxConns        int           `yaml:"max_conns"`
	MinConns        int           `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
}

type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type JWTConfig struct {
	Issuer       string        `yaml:"issuer"`
	Secret       string        `yaml:"secret"`
	AccessTTL    time.Duration `yaml:"access_ttl"`
	RefreshTTL   time.Duration `yaml:"refresh_ttl"`
	CookieSecure bool          `yaml:"cookie_secure"`
}
