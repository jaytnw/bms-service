package config

import (
	"os"
	"strconv"
)

// Config โครงสร้างการตั้งค่าของแอปพลิเคชัน
type Config struct {
	ServerAddress  string
	PostgresConfig PostgresConfig
	RedisConfig    RedisConfig
	MQTTConfig     MQTTConfig
}

// PostgresConfig โครงสร้างการตั้งค่าสำหรับ PostgreSQL
type PostgresConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	MaxIdleConns int
	MaxOpenConns int
}

// RedisConfig โครงสร้างการตั้งค่าสำหรับ Redis
type RedisConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
	PoolSize int
}

type MQTTConfig struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
}

// LoadConfig โหลดการตั้งค่าจากตัวแปรสภาพแวดล้อม
func LoadConfig() *Config {
	// ตั้งค่าเริ่มต้นสำหรับ PostgreSQL
	postgresConfig := PostgresConfig{
		Host:         getEnv("POSTGRES_HOST", "localhost"),
		Port:         getEnv("POSTGRES_PORT", "5432"),
		User:         getEnv("POSTGRES_USER", "postgres"),
		Password:     getEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:       getEnv("POSTGRES_DB", "myapp"),
		MaxIdleConns: getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 10),
		MaxOpenConns: getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 30),
	}

	// ตั้งค่าเริ่มต้นสำหรับ Redis
	redisConfig := RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Username: getEnv("REDIS_USERNAME", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
		PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
	}

	mqttConfig := MQTTConfig{
		BrokerURL: getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		ClientID:  getEnv("MQTT_CLIENT_ID", "bms-client"),
		Username:  getEnv("MQTT_USERNAME", ""),
		Password:  getEnv("MQTT_PASSWORD", ""),
	}

	return &Config{
		ServerAddress:  getEnv("SERVER_ADDRESS", ":8080"),
		PostgresConfig: postgresConfig,
		RedisConfig:    redisConfig,
		MQTTConfig:     mqttConfig,
	}
}

// getEnv ดึงค่าจากตัวแปรสภาพแวดล้อม หากไม่พบจะใช้ค่าเริ่มต้น
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt แปลงค่าจากตัวแปรสภาพแวดล้อมเป็นตัวเลข หากไม่พบจะใช้ค่าเริ่มต้น
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (p PostgresConfig) BuildDSN() string {
	return "host=" + p.Host +
		" user=" + p.User +
		" password=" + p.Password +
		" dbname=" + p.DBName +
		" port=" + p.Port +
		" sslmode=disable"
}
