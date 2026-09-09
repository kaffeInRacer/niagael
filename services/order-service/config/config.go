package config

import (
	"fmt"
	"kaffein/order-service/utils/constants"
	"os"
	"strings"

	"kaffein/order-service/utils/env"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	st := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrConfigFileNotFound, path)
	}

	if err := yaml.Unmarshal(data, st); err != nil {
		return nil, fmt.Errorf(constants.ErrConfigParse, path, err)
	}

	st.HTTP.ServiceName = env.GetString("SERVICE_NAME", st.HTTP.ServiceName)
	st.HTTP.Mode = env.GetString("HTTP_MODE", st.HTTP.Mode)
	st.HTTP.Addr = env.GetString("HTTP_ADDR", st.HTTP.Addr)
	st.HTTP.ReadTimeout = env.GetDuration("HTTP_READ_TIMEOUT", st.HTTP.ReadTimeout)
	st.HTTP.IdleTimeout = env.GetDuration("HTTP_IDLE_TIMEOUT", st.HTTP.IdleTimeout)
	st.HTTP.WriteTimeout = env.GetDuration("HTTP_WRITE_TIMEOUT", st.HTTP.WriteTimeout)
	st.HTTP.ShutdownTimeout = env.GetDuration("HTTP_SHUTDOWN_TIMEOUT", st.HTTP.ShutdownTimeout)
	st.GRPC.ProductServer = env.GetString("PRODUCT_GRPC_SERVER", st.GRPC.ProductServer)
	st.GRPC.DynamicPricingServer = env.GetString("DYNAMIC_PRICING_GRPC_SERVER", st.GRPC.DynamicPricingServer)

	st.Logger.Level = env.GetString("LOG_LEVEL", st.Logger.Level)
	st.Logger.Format = env.GetString("LOG_FORMAT", st.Logger.Format)
	st.Logger.Stdout = env.GetBool("LOG_STDOUT", st.Logger.Stdout)
	st.Logger.Path = env.GetString("LOG_PATH", st.Logger.Path)

	st.Postgres.DSN = env.GetString("POSTGRES_DSN", st.Postgres.DSN)
	st.Postgres.MaxConns = env.GetInt("POSTGRES_MAX_CONNS", st.Postgres.MaxConns)
	st.Postgres.MinConns = env.GetInt("POSTGRES_MIN_CONNS", st.Postgres.MinConns)
	st.Postgres.MaxConnLifetime = env.GetDuration("POSTGRES_MAX_CONN_LIFETIME", st.Postgres.MaxConnLifetime)
	st.Postgres.MaxConnIdleTime = env.GetDuration("POSTGRES_MAX_CONN_IDLE_TIME", st.Postgres.MaxConnIdleTime)

	st.Redis.Addr = env.GetString("REDIS_ADDR", st.Redis.Addr)
	st.Redis.Password = env.GetString("REDIS_PASSWORD", st.Redis.Password)
	st.Redis.DB = env.GetInt("REDIS_DB", st.Redis.DB)
	st.Redis.PoolSize = env.GetInt("REDIS_POOL_SIZE", st.Redis.PoolSize)
	st.Redis.MinIdleConns = env.GetInt("REDIS_MIN_IDLE_CONNS", st.Redis.MinIdleConns)
	st.Redis.MaxConnLifetime = env.GetDuration("REDIS_MAX_CONN_LIFETIME", st.Redis.MaxConnLifetime)
	st.Redis.MaxConnIdleTime = env.GetDuration("REDIS_MAX_CONN_IDLE_TIME", st.Redis.MaxConnIdleTime)

	st.Midtrans.ServerKey = env.GetString("MIDTRANS_SERVER_KEY", st.Midtrans.ServerKey)
	st.Midtrans.ClientKey = env.GetString("MIDTRANS_CLIENT_KEY", st.Midtrans.ClientKey)
	st.Midtrans.Environment = env.GetString("MIDTRANS_ENVIRONMENT", st.Midtrans.Environment)
	st.JWT.Secret = env.GetString("JWT_SECRET", st.JWT.Secret)
	st.JWT.Issuer = env.GetString("JWT_ISSUER", st.JWT.Issuer)
	st.JWT.RedisDB = env.GetInt("JWT_REDIS_DB", st.JWT.RedisDB)
	st.RBAC.PolicyReloadInterval = env.GetDuration("RBAC_POLICY_RELOAD_INTERVAL", st.RBAC.PolicyReloadInterval)
	st.Kafka.Brokers = env.GetStringSlice("KAFKA_BROKERS", st.Kafka.Brokers)
	st.Kafka.GroupID = env.GetString("KAFKA_GROUP_ID", st.Kafka.GroupID)
	st.Kafka.CasbinTopic = env.GetString("KAFKA_CASBIN_TOPIC", st.Kafka.CasbinTopic)
	st.Kafka.OrderTopic = env.GetString("KAFKA_ORDER_TOPIC", st.Kafka.OrderTopic)
	st.Kafka.UserTopic = env.GetString("KAFKA_USER_TOPIC", st.Kafka.UserTopic)
	if strings.TrimSpace(st.JWT.Secret) == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if !strings.EqualFold(st.HTTP.Mode, "development") &&
		(strings.TrimSpace(st.Midtrans.ServerKey) == "" || strings.TrimSpace(st.Midtrans.ClientKey) == "") {
		return nil, fmt.Errorf("MIDTRANS_SERVER_KEY and MIDTRANS_CLIENT_KEY are required outside development")
	}

	return st, nil
}
