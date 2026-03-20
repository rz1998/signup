package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Database  DatabaseConfig    `json:"database"`
	Cache     CacheConfig      `json:"cache"`
	JWT       JWTConfig        `json:"jwt"`
	CORS      CORSConfig       `json:"cors"`
	RateLimit RateLimitConfig  `json:"rateLimit"`
	BaseURL   string           `json:"baseURL"`
}

type DatabaseConfig struct {
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DBName       string `json:"dbName"`
	User         string `json:"user"`
	Password     string `json:"password"`
	SSLMode      string `json:"sslMode"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxIdleConns int    `json:"maxIdleConns"`
}

type CacheConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type JWTConfig struct {
	Secret   string `json:"secret"`
	Expire   int    `json:"expire"`
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

type CORSConfig struct {
	AllowOrigins []string `json:"allowOrigins"`
	AllowMethods []string `json:"allowMethods"`
	AllowHeaders []string `json:"allowHeaders"`
}

type RateLimitConfig struct {
	Enabled bool `json:"enabled"`
	QPS     int  `json:"qps"`
}
