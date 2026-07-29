package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv          string `mapstructure:"APP_ENV"`
	Port            string `mapstructure:"PORT"`
	DBHost          string `mapstructure:"DB_HOST"`
	DBPort          string `mapstructure:"DB_PORT"`
	DBUser          string `mapstructure:"DB_USER"`
	DBPass          string `mapstructure:"DB_PASS"`
	DBName          string `mapstructure:"DB_NAME"`
	RedisAddr       string `mapstructure:"REDIS_ADDR"`
	RedisPass       string `mapstructure:"REDIS_PASS"`
	RabbitMQURL     string `mapstructure:"RABBITMQ_URL"`
	CacheTTLMinutes int    `mapstructure:"CACHE_TTL_MINUTES"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8084")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("CACHE_TTL_MINUTES", 15)
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found for analytics-service, using environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
