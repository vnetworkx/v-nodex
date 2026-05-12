package config

import "os"

type Config struct {
	DBDSN     string
	RedisAddr string
	NodeAddr  string
	NodeID    string
}

func FromEnv() Config {
	return Config{
		DBDSN:     getenv("VNODEX_DB_DSN", "postgres://vnode:vnode@127.0.0.1:5432/vnode?sslmode=disable"),
		RedisAddr: getenv("VNODEX_REDIS_ADDR", "127.0.0.1:6379"),
		NodeAddr:  getenv("VNODEX_NODE_ADDR", ":9090"),
		NodeID:    getenv("VNODEX_NODE_ID", "node-local"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
