package config

type RedisConfig struct {
	Host string
	Port string
}

func LoadRedisConfig() (*RedisConfig, error) {
	return &RedisConfig{
		Host: getEnv("REDIS_HOST", "localhost"),
		Port: getEnv("REDIS_PORT", "6379"),
	}, nil
}
