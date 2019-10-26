package apitest

import (
	"github.com/spf13/viper"
	"gopkg.in/h2non/baloo.v3"
)

var test *baloo.Client

func setupConfiguration() {
	_ = viper.BindEnv("minio.addr", "MINIO_ADDR")
	_ = viper.BindEnv("minio.access", "MINIO_ACCESS_KEY")
	_ = viper.BindEnv("minio.secret", "MINIO_SECRET_KEY")
	viper.SetDefault("minio.addr", "localhost:9000")
	viper.SetDefault("minio.access", "minio")
	viper.SetDefault("minio.secret", "insecure")

	_ = viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.SetDefault("redis.addr", "localhost:6379")

	_ = viper.BindEnv("api.URL", "API_URL")
	viper.SetDefault("api.URL", "http://localhost:8080")
}

func init() {
	setupConfiguration()
	// test stores the HTTP testing client preconfigured
	test = baloo.New(viper.GetString("api.URL"))
}
