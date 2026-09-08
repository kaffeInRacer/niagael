package config

import "time"

type Config struct {
	HTTP     HttpConfig     `yaml:"http"`
	GRPC     GrpcConfig     `yaml:"grpc"`
	Logger   LoggerConfig   `yaml:"log"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Minio    MinIOConfig    `yaml:"minio"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type JWTConfig struct {
	Secret  string `yaml:"secret"`
	Issuer  string `yaml:"issuer"`
	RedisDB int    `yaml:"redis_db"`
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
	Server         string `yaml:"server"`
	DynamicPricing string `yaml:"dynamic_pricing"`
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

type KafkaConfig struct {
	Brokers []string          `yaml:"brokers"`
	GroupID string            `yaml:"group_id"`
	Topics  map[string]string `yaml:"topics"`
}

type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
	UseSSL          bool   `yaml:"use_ssl"`
}
