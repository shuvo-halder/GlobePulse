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
	JWTSecret       string `mapstructure:"JWT_SECRET"`
	TokenExpiration int    `mapstructure:"TOKEN_EXPIRATION_MINUTES"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8081")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("TOKEN_EXPIRATION_MINUTES", 60)
	viper.SetDefault("JWT_SECRET", "super-secret-key-change-in-production")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found for auth-service, using environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
