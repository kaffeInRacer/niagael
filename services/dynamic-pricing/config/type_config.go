package config

import "time"

type Config struct {
	HTTP     HttpConfig     `yaml:"http"`
	GRPC     GrpcConfig     `yaml:"grpc"`
	Logger   LoggerConfig   `yaml:"log"`
	Postgres PostgresConfig `yaml:"postgres"`
	AuthDB   PostgresConfig `yaml:"auth_db"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	RBAC     RBACConfig     `yaml:"rbac"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type KafkaConfig struct {
	Brokers     []string `yaml:"brokers"`
	GroupID     string   `yaml:"group_id"`
	CasbinTopic string   `yaml:"casbin_topic"`
}

type JWTConfig struct {
	Secret  string `yaml:"secret"`
	Issuer  string `yaml:"issuer"`
	RedisDB int    `yaml:"redis_db"`
}

type RBACConfig struct {
	PolicyReloadInterval time.Duration `yaml:"policy_reload_interval"`
}

type HttpConfig struct {
	ServiceName     string        `yaml:"service_name"`
	Mode            string        `yaml:"mode"`
	Addr            string        `yaml:"addr"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type GrpcConfig struct {
	Server string `yaml:"server"`
}

type PostgresConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxConns        int           `yaml:"max_conns"`
	MinConns        int           `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
}

type RedisConfig struct {
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Stdout bool   `yaml:"stdout"`
	Path   string `yaml:"path"`
}
