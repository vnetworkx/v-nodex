package config

import "os"

type Config struct {
	DBDSN     string
	RedisAddr string
	HTTPAddr  string
}

func FromEnv() Config {
	cfg := Config{
		DBDSN:     getenv("VNODEX_DB_DSN", "postgres://vnode:vnode@127.0.0.1:5432/vnode?sslmode=disable"),
		RedisAddr: getenv("VNODEX_REDIS_ADDR", "127.0.0.1:6379"),
		HTTPAddr:  getenv("VNODEX_HTTP_ADDR", ":8080"),
	}
	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
